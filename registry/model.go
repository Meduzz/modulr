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
		Lookup(string) ([]api.Service, error)
		// Plugin - register a lifecycle plugin
		Plugin(Lifecycle)
		// Start - tell the service registry to cold start
		Start() error
		// SetStorage - set the storage to be used by this registry
		SetStorage(RegistryStorage)
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

	// RegistryStorage - storage adapter for stuff in the registry
	RegistryStorage interface {
		// Store - store a service by its name and id
		Store(string, api.Service) error
		// Remove - remove a service by its name and id
		Remove(string, string) (api.Service, error)
		// Lookup - fetch all instance of service by its name
		Lookup(string) ([]api.Service, error)
		// Start - tell the storage to cold start
		Start() error
	}
)
