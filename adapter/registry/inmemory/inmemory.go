package inmemory

import (
	"github.com/Meduzz/modulr"
	"github.com/Meduzz/modulr/api"
)

type (
	inmemoryStorage struct {
		services map[string][]api.Service // name -> services
	}
)

func init() {
	modulr.ServiceRegistry.SetStorage(NewInMemoryStorage())
}

// NewInMemoryStorage - stores stuff in memory, do not use :)
func NewInMemoryStorage() api.RegistryStorage {
	svcs := make(map[string][]api.Service)
	return &inmemoryStorage{svcs}
}

// Store - store a service by its name and id
func (i *inmemoryStorage) Store(name string, service api.Service) error {
	services, exists := i.services[name]

	if !exists {
		services = make([]api.Service, 0)
	}

	services = append(services, service)
	i.services[name] = services

	return nil
}

// Remove - remove a service by its name and id
func (i *inmemoryStorage) Remove(name string, id string) (api.Service, error) {
	it, exists := i.services[name]

	if !exists {
		return nil, nil
	}

	var removed api.Service

	keepers := make([]api.Service, 0)

	for _, service := range it {
		if service.GetID() != id {
			keepers = append(keepers, service)
		} else {
			removed = service
		}
	}

	if len(keepers) == 0 {
		delete(i.services, name)
	} else {
		i.services[name] = keepers
	}

	return removed, nil
}

// Lookup - fetch all instance of service by its name
func (i *inmemoryStorage) Lookup(name string) ([]api.Service, error) {
	it, exists := i.services[name]

	if !exists {
		return make([]api.Service, 0), nil
	}

	return it, nil
}

// Start - tell the storage to cold start
func (i *inmemoryStorage) Start() ([]string, error) {
	return nil, nil
}
