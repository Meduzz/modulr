package registry

import (
	"github.com/Meduzz/modulr/api"
)

type (
	// ServiceRegistry - provids main api for the framework.
	ServiceRegistry interface {
		// Register - register a service
		Register(*api.Service)
		// Deregister  - remove a service by id
		Deregister(string)
	}

	// Lifecycle - provides lifecycle methods for child modules.
	Lifecycle interface {
		// Register - a new service was added to the registry
		Register(*api.Service)
		// Deregister - a service was removed from the registry
		Deregister(*api.Service)
	}

	serviceRegistry struct {
		services map[string]*api.Service // id -> services
		children []Lifecycle             // "child" modules
	}
)

// NewServiceRegistry - creates a new service registry with any provided children
func NewServiceRegistry(children ...Lifecycle) ServiceRegistry {
	svcs := make(map[string]*api.Service)

	return &serviceRegistry{
		services: svcs,
		children: children,
	}
}

func (s *serviceRegistry) Register(service *api.Service) {
	_, exists := s.services[service.ID]

	if !exists {
		s.services[service.ID] = service

		for _, child := range s.children {
			child.Register(service)
		}
	}
}

func (s *serviceRegistry) Deregister(id string) {
	it, exists := s.services[id]

	if !exists {
		return
	}

	defer delete(s.services, id)

	for _, child := range s.children {
		child.Deregister(it)
	}
}
