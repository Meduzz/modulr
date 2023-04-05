package event

import (
	"log"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/errorz"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/Meduzz/modulr/registry"
)

type (
	subscriptionRegistry struct {
		adapter          EventAdapter
		deliveryAdapters map[string]Deliverer
		register         registry.ServiceRegistry
		factory          loadbalancer.LoadBalancerFactory
	}
)

// NewEventSupport - creates a new EventSupport with the provided adapter
func NewEventSupport(register registry.ServiceRegistry,
	eventAdapter EventAdapter,
	factory loadbalancer.LoadBalancerFactory) EventDelivery {

	sub := &subscriptionRegistry{
		adapter:          eventAdapter,
		deliveryAdapters: make(map[string]Deliverer),
		factory:          factory,
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

func (s *subscriptionRegistry) RegisterDeliverer(serviceType string, adapter Deliverer) {
	s.deliveryAdapters[serviceType] = adapter
}

func (s *subscriptionRegistry) eventHandler(name string, sub *api.Subscription) func([]byte) {
	return func(body []byte) {
		services := s.register.Lookup(name)
		lb := s.factory.For(name)
		service := lb.Next(services)

		if service == nil {
			log.Printf("Loadbalancer returned nil service (%s)\n", name)
			// TODO safe to unsubscribe?
			return
		}

		err := s.deliveryAdapters[service.GetType()].DeliverEvent(service, sub, body)

		if err != nil {
			// TODO do something smarter with errors
			log.Printf("Delivering event to %s threw error: %v\n", sub.Path, err)
		} else {
			log.Printf("Delivering event to %s went well.", sub.Path)
		}
	}
}
