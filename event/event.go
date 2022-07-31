package event

import (
	"fmt"
	"log"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/errorz"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/Meduzz/modulr/registry"
)

type (
	subscriptionRegistry struct {
		adapter         EventAdapter
		deliveryAdapter DeliveryAdapter
		register        registry.ServiceRegistry
		factory         loadbalancer.LoadBalancerFactory
	}
)

// NewEventSupport - creates a new EventSupport with the provided adapter
func NewEventSupport(register registry.ServiceRegistry,
	eventAdapter EventAdapter,
	deliveryAdapter DeliveryAdapter,
	factory loadbalancer.LoadBalancerFactory) registry.Lifecycle {

	sub := &subscriptionRegistry{
		adapter:         eventAdapter,
		deliveryAdapter: deliveryAdapter,
		factory:         factory,
		register:        register,
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

func (s *subscriptionRegistry) eventHandler(name string, sub *api.Subscription) func([]byte) {
	return func(body []byte) {
		services := s.register.Lookup(name)
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
