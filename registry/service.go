package registry

import (
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/errorz"
)

type (
	// ServiceRegistry - provids main api for the framework.
	ServiceRegistry interface {
		// Register - register a service
		Register(*api.Service) error
		// Deregister  - remove a service by id
		Deregister(string) error
	}

	// Lifecycle - provides lifecycle methods for child modules.
	Lifecycle interface {
		// Register - a new service was added to the registry
		Register(*api.Service) error
		// Deregister - a service was removed from the registry
		Deregister(*api.Service) error
	}

	serviceRegistry struct {
		services map[string]*api.Service // id -> services
		children []Lifecycle             // "child" modules
	}
)

// NewServiceRegistry - creates a new service registry
func NewServiceRegistry(children ...Lifecycle) ServiceRegistry {
	return &serviceRegistry{
		services: make(map[string]*api.Service),
		children: children,
	}
}

func (s *serviceRegistry) Register(service *api.Service) error {
	_, exists := s.services[service.ID]

	if !exists {
		s.services[service.ID] = service
		combined := errorz.NewError(nil)

		for _, child := range s.children {
			err := child.Register(service)
			combined.Append(err)
		}

		return combined.Error()
	}

	return nil
}

func (s *serviceRegistry) Deregister(id string) error {
	it, exists := s.services[id]

	if !exists {
		return nil
	}

	combined := errorz.NewError(nil)
	defer delete(s.services, id)

	for _, child := range s.children {
		err := child.Deregister(it)
		combined.Append(err)
	}

	return combined.Error()
}
