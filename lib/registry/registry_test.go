package registry

import (
	"fmt"
	"testing"

	"github.com/Meduzz/modulr/api"
)

type (
	plugin struct{}

	storage struct {
		svcs []api.Service
	}
)

var (
	pluginError  bool
	storageError bool
	serviceName  chan string
	subject      *serviceRegistry
	service1     api.Service
	service2     api.Service
)

func TestMain(m *testing.M) {
	pluginError = false
	storageError = false
	serviceName = make(chan string, 10)

	subject = &serviceRegistry{
		children: make([]api.Lifecycle, 0),
	}

	subject.Plugin(NewPlugin())
	subject.SetStorage(NewStorage())

	service1 = &api.DefaultService{
		ID:   "1",
		Name: "test",
	}

	service2 = &api.DefaultService{
		ID:   "2",
		Name: "test",
	}

	m.Run()
}

func TestHappyPath(t *testing.T) {
	err := subject.Register(service1)

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName
	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	err = subject.Register(service2)

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	svcs, err := subject.Lookup("test")

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	if len(svcs) != 2 {
		t.Errorf("expected number of registered services to be 2 but was %d", len(svcs))
	}

	err = subject.Deregister(service1.GetName(), service1.GetID())

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	err = subject.Deregister(service2.GetName(), service2.GetID())

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName
	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	svcs, err = subject.Lookup("test")

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	if len(svcs) > 0 {
		t.Errorf("expected number of registered services to be 0 but was %d", len(svcs))
	}
}

func TestAddSameTwice(t *testing.T) {
	err := subject.Register(service1)

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName
	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	err = subject.Register(service1)

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}
}

func TestRemoveSameTwice(t *testing.T) {
	err := subject.Deregister(service1.GetName(), service1.GetID())

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	<-serviceName
	<-serviceName

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}

	err = subject.Deregister(service1.GetName(), service1.GetID())

	if err != nil {
		t.Errorf("There was an unexpected error: %v", err)
	}

	if len(serviceName) > 0 {
		t.Error("there were too many calls to the plugin")
	}
}

// let storage implement RegistryStorage
func NewStorage() api.RegistryStorage {
	return &storage{make([]api.Service, 0)}
}

func (s *storage) Store(name string, svc api.Service) error {
	if storageError {
		return fmt.Errorf("im an error")
	}

	s.svcs = append(s.svcs, svc)

	return nil
}

func (s *storage) Remove(name, id string) (api.Service, error) {
	if storageError {
		return nil, fmt.Errorf("im an error")
	}

	keepers := make([]api.Service, 0)
	var removed api.Service

	for _, it := range s.svcs {
		if it.GetID() != id {
			keepers = append(keepers, it)
		} else {
			removed = it
		}
	}

	s.svcs = keepers

	return removed, nil
}

func (s *storage) Lookup(name string) ([]api.Service, error) {
	if storageError {
		return nil, fmt.Errorf("im an error")
	}

	return s.svcs, nil
}

func (s *storage) Start() ([]string, error) {
	if storageError {
		return nil, fmt.Errorf("im an error")
	}

	return nil, nil
}

// let plugin implement Lifecycle
func NewPlugin() api.Lifecycle {
	return &plugin{}
}

func (p *plugin) RegisterService(svc api.Service) error {
	if pluginError {
		return fmt.Errorf("im an error")
	}

	serviceName <- svc.GetName()

	return nil
}

func (p *plugin) DeregisterService(svc api.Service) error {
	if pluginError {
		return fmt.Errorf("im an error")
	}

	serviceName <- svc.GetName()

	return nil
}

func (p *plugin) RegisterInstance(svc api.Service) error {
	if pluginError {
		return fmt.Errorf("im an error")
	}

	serviceName <- svc.GetName()

	return nil
}

func (p *plugin) DeregisterInstance(svc api.Service) error {
	if pluginError {
		return fmt.Errorf("im an error")
	}

	serviceName <- svc.GetName()

	return nil
}
