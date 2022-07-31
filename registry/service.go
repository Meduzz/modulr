package registry

import (
	"github.com/Meduzz/modulr/api"
)

type (
	// ServiceRegistry - provids main api for the framework.
	ServiceRegistry interface {
		// Register - register a service
		Register(api.Service) error
		// Deregister - remove a service by name & id
		Deregister(string, string) error
		// Lookup - fetch services by name, never null
		Lookup(string) []api.Service
		// Plugin - register a lifecycle plugin
		Plugin(Lifecycle)
		// Start - tell the service registry to cold start
		Start() error
	}

	// Lifecycle - provides lifecycle methods for child modules.
	Lifecycle interface {
		// RegisterService - a new service was created in the registry
		RegisterService(api.Service) error
		// DeregisterService - a service was completely removed from the registry
		DeregisterService(api.Service) error
		// RegisterInstance - one service instance was added to the registry
		RegisterInstance(api.Service) error
		// DeregisterInstance - one service instance was removed from the registry
		DeregisterInstance(api.Service) error
	}

	serviceRegistry struct {
		services map[string][]api.Service // name -> services
		children []Lifecycle
	}
)

// NewServiceRegistry - creates a new in memory service registry
func NewServiceRegistry() ServiceRegistry {
	registry := &serviceRegistry{
		services: make(map[string][]api.Service),
		children: make([]Lifecycle, 0),
	}

	return registry
}

func (s *serviceRegistry) Register(service api.Service) error {
	services, exists := s.services[service.GetName()]

	if !exists {
		for _, child := range s.children {
			child.RegisterService(service)
		}

		services = make([]api.Service, 0)
	}

	services = append(services, service)
	s.services[service.GetName()] = services

	for _, child := range s.children {
		child.RegisterInstance(service)
	}

	return nil
}

func (s *serviceRegistry) Deregister(name, id string) error {
	it, exists := s.services[name]

	if !exists {
		return nil
	}

	keepers := make([]api.Service, 0)
	var old api.Service

	for _, service := range it {
		if service.GetID() != id {
			keepers = append(keepers, service)
		} else {
			old = service
			for _, child := range s.children {
				child.DeregisterInstance(service)
			}
		}
	}

	if len(keepers) == 0 {
		for _, child := range s.children {
			child.DeregisterService(old)
		}

		delete(s.services, name)
		return nil
	}

	s.services[name] = keepers

	return nil
}

func (s *serviceRegistry) Lookup(name string) []api.Service {
	it, exists := s.services[name]

	if !exists {
		return make([]api.Service, 0)
	}

	return it
}

func (s *serviceRegistry) Plugin(lc Lifecycle) {
	s.children = append(s.children, lc)
}

func (s *serviceRegistry) Start() error {
	return nil
}
