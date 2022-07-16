package registry

import (
	"fmt"
	"log"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/errorz"
	"github.com/Meduzz/modulr/event"
	"github.com/Meduzz/modulr/loadbalancer"
)

type (
	// EventProxy - api to proxy events from and to http
	EventProxy interface {
		Lifecycle
		// Publish - api to publish an event
		Publish(*api.Event) error
		// Request - api to rpc over event adapter
		Request(*api.Event, string) ([]byte, error)
	}

	subscriptionRegistry struct {
		adapter         event.EventAdapter
		deliveryAdapter event.DeliveryAdapter
		registry        ServiceRegistry
		factory         loadbalancer.LoadBalancerFactory
	}
)

// NewEventProxy - creates a new EventRegistry with the provided adapter
func NewEventProxy(registry ServiceRegistry,
	eventAdapter event.EventAdapter,
	deliveryAdapter event.DeliveryAdapter,
	factory loadbalancer.LoadBalancerFactory) EventProxy {

	sub := &subscriptionRegistry{
		adapter:         eventAdapter,
		deliveryAdapter: deliveryAdapter,
		factory:         factory,
		registry:        registry,
	}

	registry.Plugin(sub)

	return sub
}

func (s *subscriptionRegistry) RegisterService(name string, service api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.GetSubscriptions() {
		err := s.adapter.Subscribe(sub.Topic, sub.Routing, sub.Group, s.eventHandler(name, sub))

		if err != nil {
			combined.Append(err)
		}
	}

	return combined.Error()
}

func (s *subscriptionRegistry) DeregisterService(name string, service api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.GetSubscriptions() {
		err := s.adapter.Unsubscribe(sub.Topic, sub.Routing, sub.Group)

		if err != nil {
			combined.Append(err)
		}
	}

	return combined.Error()
}

func (s *subscriptionRegistry) RegisterInstance(service api.Service) error {
	return nil
}

func (s *subscriptionRegistry) DeregisterInstance(service api.Service) error {
	return nil
}

func (s *subscriptionRegistry) Publish(event *api.Event) error {
	return s.adapter.Publish(event.Topic, event.Routing, event.Body)
}

func (s *subscriptionRegistry) Request(event *api.Event, maxWait string) ([]byte, error) {
	return s.adapter.Request(event.Topic, event.Routing, event.Body, maxWait)
}

func (s *subscriptionRegistry) eventHandler(name string, sub *api.Subscription) func([]byte) {
	return func(body []byte) {
		services := s.registry.Lookup(name)
		lb := s.factory.For(name)
		service := lb.Next(services)

		if service == nil {
			log.Printf("Loadbalancer returned nil service (%s)\n", name)
			return
		}

		// TODO assuming http
		url := fmt.Sprintf("http://%s:%d%s%s", service.GetAddress(), service.GetPort(), service.GetContext(), sub.Path)
		err := s.deliveryAdapter.DeliverEvent(url, body)

		if err != nil {
			// TODO do something smarter with errors
			log.Printf("Delivering event to %s threw error: %v\n", url, err)
		}
	}
}
