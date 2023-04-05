package event

import (
	"log"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/lib/errorz"
)

type (
	subscriptionRegistry struct {
		adapter          api.EventAdapter
		deliveryAdapters map[string]api.EventDeliveryAdapter
		register         api.ServiceRegistry
		factory          api.LoadBalancerFactory
	}
)

// NewEventSupport - creates a new EventSupport with the provided adapter
func NewEventSupport(register api.ServiceRegistry) api.EventSupport {
	sub := &subscriptionRegistry{
		deliveryAdapters: make(map[string]api.EventDeliveryAdapter),
		register:         register,
	}

	register.Plugin(sub)

	return sub
}

func (s *subscriptionRegistry) RegisterService(service api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.GetSubscriptions() {
		err := s.adapter.Subscribe(sub.Topic, sub.Routing, sub.Group, s.eventHandler(service.GetName(), sub))

		if err != nil {
			combined.Append(err)
		}
	}

	return combined.Error()
}

func (s *subscriptionRegistry) DeregisterService(service api.Service) error {
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

func (s *subscriptionRegistry) RegisterDeliverer(serviceType string, adapter api.EventDeliveryAdapter) {
	s.deliveryAdapters[serviceType] = adapter
}

func (s *subscriptionRegistry) SetEventAdapter(adapter api.EventAdapter) {
	s.adapter = adapter
}

func (s *subscriptionRegistry) Publish(event *api.Event) error {
	return s.adapter.Publish(event.Topic, event.Routing, event.Body)
}

func (s *subscriptionRegistry) Request(event *api.Event, maxWait string) ([]byte, error) {
	return s.adapter.Request(event.Topic, event.Routing, event.Body, maxWait)
}

func (s *subscriptionRegistry) SetLoadBalancerFactory(factory api.LoadBalancerFactory) {
	s.factory = factory
}

func (s *subscriptionRegistry) eventHandler(name string, sub *api.Subscription) func([]byte) {
	return func(body []byte) {
		services, err := s.register.Lookup(name)

		if err != nil {
			// TODO do something smarter with errors
			log.Printf("Looking up services for service %s threw error: %v\n", name, err)
		}

		lb := s.factory.For(name)
		service := lb.Next(services)

		if service == nil {
			log.Printf("Loadbalancer returned nil service (%s)\n", name)
			// TODO safe to unsubscribe?
			return
		}

		err = s.deliveryAdapters[service.GetType()].Deliver(service, sub, body)

		if err != nil {
			// TODO do something smarter with errors
			log.Printf("Delivering event to %s threw error: %v\n", sub.Path, err)
		} else {
			log.Printf("Delivering event to %s went well.", sub.Path)
		}
	}
}
