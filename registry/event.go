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
	// EventRegistry - event module api
	EventRegistry interface {
		Lifecycle
		// Publish - api to publish an event
		Publish(*api.Event) error
	}

	subscriptionRegistry struct {
		adapter         event.EventAdapter
		deliveryAdapter event.DeliveryAdapter
		subscriptions   map[string]loadbalancer.LoadBalancer // name -> loadbalancer
		registry        ServiceRegistry
		factory         loadbalancer.LoadBalancerFactory
	}
)

// NewEventRegistry - creates a new EventRegistry with the provided adapter
func NewEventRegistry(eventAdapter event.EventAdapter,
	deliveryAdapter event.DeliveryAdapter,
	factory loadbalancer.LoadBalancerFactory) EventRegistry {

	subs := make(map[string]loadbalancer.LoadBalancer)

	return &subscriptionRegistry{
		adapter:         eventAdapter,
		deliveryAdapter: deliveryAdapter,
		subscriptions:   subs,
		factory:         factory,
	}
}

func (s *subscriptionRegistry) RegisterService(name string, service api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.GetSubscriptions() {
		lb := s.factory.Create()
		err := s.adapter.Subscribe(sub.Topic, sub.Routing, sub.Group, s.eventHandler(name, sub, lb))

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

func (s *subscriptionRegistry) ServiceRegistry(registry ServiceRegistry) {
	s.registry = registry
}

func (s *subscriptionRegistry) eventHandler(name string, sub *api.Subscription, lb loadbalancer.LoadBalancer) func([]byte) {
	return func(body []byte) {
		services := s.registry.Lookup(name)
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
