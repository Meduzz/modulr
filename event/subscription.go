package event

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/registry"
)

type (
	// EventRegistry - event module api
	EventRegistry interface {
		registry.Lifecycle
		// Publish - api to publish an event
		Publish(*api.Event)
	}

	// EventAdapter - interface to be implemented by event adapters
	EventAdapter interface {
		// Publish - publish a payload to a topic with optional routingkey
		Publish(string, string, []byte)
		// Subscribe - subscribe to a topic with optional routingkey and subscribergroup
		Subscribe(string, string, string, func([]byte))
		// Unsubscribe - unsubscribe to a topic with optional routingkey and subscribergroup
		Unsubscribe(string, string, string)
	}

	subscriptionRegistry struct {
		adapter       EventAdapter
		subscriptions map[string]*subscription // topic.routing.group -> subscription
	}
)

// NewEventRegistry - creates a new EventRegistry with the provided adapter
func NewEventRegistry(adapter EventAdapter) EventRegistry {
	subs := make(map[string]*subscription)

	return &subscriptionRegistry{
		adapter:       adapter,
		subscriptions: subs,
	}
}

func (s *subscriptionRegistry) Register(service *api.Service) {
	for _, sub := range service.Subscriptions {
		current := s.upsertSubscription(sub)

		exists := false

		for _, existing := range current.Services {
			if existing.ID == service.ID {
				exists = true
			}
		}

		if !exists {
			svc := &subscribee{
				ID:      service.ID,
				Address: service.Address,
				Port:    service.Port,
				Context: service.Context,
				Path:    sub.Path,
			}

			current.Services = append(current.Services, svc)
		}
	}
}

func (s *subscriptionRegistry) Deregister(service *api.Service) {
	for _, sub := range service.Subscriptions {
		current := s.upsertSubscription(sub)
		copy := make([]*subscribee, 0)

		for _, active := range current.Services {
			if active.ID != service.ID {
				copy = append(copy, active)
			}
		}

		current.Services = copy

		if len(current.Services) == 0 {
			s.adapter.Unsubscribe(sub.Topic, sub.Routing, sub.Group)
		}
	}
}

func (s *subscriptionRegistry) Publish(event *api.Event) {
	s.adapter.Publish(event.Topic, event.Routing, event.Body)
}

func (s *subscriptionRegistry) upsertSubscription(sub *api.Subscription) *subscription {
	key := fmt.Sprintf("%s.%s.%s", sub.Topic, sub.Routing, sub.Group)

	it, exists := s.subscriptions[key]

	if !exists {
		it = &subscription{
			Topic:    sub.Topic,
			Routing:  sub.Routing,
			Group:    sub.Group,
			Services: make([]*subscribee, 0),
		}

		s.adapter.Subscribe(sub.Topic, sub.Routing, sub.Group, s.eventHandler(it))
	}

	return it
}

func (s *subscriptionRegistry) eventHandler(sub *subscription) func([]byte) {
	index := 0
	return func(body []byte) {
		if index >= len(sub.Services) {
			index = 0
		}

		service := sub.Services[index]
		index++

		url := fmt.Sprintf("http://%s:%s/%s/%s", service.Address, service.Path, service.Context, service.Path)
		req, err := client.POSTBytes(url, body)

		if err != nil {
			log.Printf("Creating request failed: %v\n", err)
			return
		}

		res, err := req.Do(http.DefaultClient)

		if err != nil {
			log.Printf("Call to %s:%d returned error: %v\n", service.Address, service.Port, err)
			return
		}

		if res.Code() != 200 {
			log.Printf("Call to %s:%d from %s did not return 200\n", service.Address, service.Path, sub.Topic)
		}
	}
}
