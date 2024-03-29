package registry

import "github.com/Meduzz/modulr/api"

type (
	serviceRegistry struct {
		children []api.Lifecycle
		storage  api.RegistryStorage
	}
)

// NewServiceRegistry - creates a new in memory service registry
func NewServiceRegistry() api.ServiceRegistry {
	registry := &serviceRegistry{
		children: make([]api.Lifecycle, 0),
	}

	return registry
}

func (s *serviceRegistry) Register(service api.Service) error {
	existing, err := s.storage.Lookup(service.GetName())

	if err != nil {
		return err
	}

	if len(existing) == 0 {
		for _, child := range s.children {
			child.RegisterService(service)
		}
	}

	err = s.storage.Store(service.GetName(), service)

	if err != nil {
		return err
	}

	for _, child := range s.children {
		child.RegisterInstance(service)
	}

	return nil
}

func (s *serviceRegistry) Deregister(name, id string) (api.Service, error) {
	svc, err := s.storage.Remove(name, id)

	if err != nil {
		return nil, err
	}

	if svc != nil {
		for _, child := range s.children {
			child.DeregisterInstance(svc)
		}

		existing, err := s.storage.Lookup(name)

		if err != nil {
			return nil, err
		}

		if len(existing) == 0 {
			for _, child := range s.children {
				child.DeregisterService(svc)
			}
		}
	}

	return svc, nil
}

func (s *serviceRegistry) Lookup(name string) ([]api.Service, error) {
	return s.storage.Lookup(name)
}

func (s *serviceRegistry) Plugin(lc api.Lifecycle) {
	s.children = append(s.children, lc)
}

func (s *serviceRegistry) Start() error {
	services, err := s.storage.Start()

	if err != nil {
		return err
	}

	for _, it := range services {
		svcs, err := s.Lookup(it)

		if err != nil {
			return err
		}

		first := true
		for _, svc := range svcs {
			for _, child := range s.children {
				if first {
					child.RegisterService(svc)
					first = false
				}

				child.RegisterInstance(svc)
			}
		}
	}

	return nil
}

func (s *serviceRegistry) SetStorage(storage api.RegistryStorage) {
	s.storage = storage
}
