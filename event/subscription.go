package event

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/errorz"
	"github.com/Meduzz/modulr/registry"
)

type (
	// EventRegistry - event module api
	EventRegistry interface {
		registry.Lifecycle
		// Publish - api to publish an event
		Publish(*api.Event) error
	}

	// EventAdapter - interface to be implemented by event adapters
	EventAdapter interface {
		// Publish - publish a payload to a topic with optional routingkey
		Publish(string, string, []byte) error
		// Subscribe - subscribe to a topic with optional routingkey and subscribergroup
		Subscribe(string, string, string, func([]byte)) error
		// Unsubscribe - unsubscribe to a topic with optional routingkey and subscribergroup
		Unsubscribe(string, string, string) error
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

func (s *subscriptionRegistry) Register(service *api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.Subscriptions {
		current, err := s.upsertSubscription(sub)

		if err != nil {
			combined.Append(err)
			continue
		}

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
			log.Printf("%s is now subscribed to topic:%s routing:%s group:%s\n", service.ID, sub.Topic, sub.Routing, sub.Group)
		}
	}

	return combined.Error()
}

func (s *subscriptionRegistry) Deregister(service *api.Service) error {
	combined := errorz.NewError(nil)

	for _, sub := range service.Subscriptions {
		current, err := s.upsertSubscription(sub) // TODO might have just created a subscription...

		if err != nil {
			combined.Append(err)
			continue
		}

		copy := make([]*subscribee, 0)

		for _, active := range current.Services {
			if active.ID != service.ID {
				copy = append(copy, active)
			}
		}

		current.Services = copy

		log.Printf("%s is now unsubscribed from topic:%s routing:%s group:%s\n", service.ID, sub.Topic, sub.Routing, sub.Group)

		if len(current.Services) == 0 {
			err = s.adapter.Unsubscribe(sub.Topic, sub.Routing, sub.Group)
			combined.Append(err)
		}
	}

	return combined.Error()
}

func (s *subscriptionRegistry) Publish(event *api.Event) error {
	return s.adapter.Publish(event.Topic, event.Routing, event.Body)
}

func (s *subscriptionRegistry) upsertSubscription(sub *api.Subscription) (*subscription, error) {
	key := fmt.Sprintf("%s.%s.%s", sub.Topic, sub.Routing, sub.Group)

	it, exists := s.subscriptions[key]

	if !exists {
		it = &subscription{
			Topic:    sub.Topic,
			Routing:  sub.Routing,
			Group:    sub.Group,
			Services: make([]*subscribee, 0),
		}

		err := s.adapter.Subscribe(sub.Topic, sub.Routing, sub.Group, s.eventHandler(it))

		if err != nil {
			return nil, err
		}

		s.subscriptions[key] = it
	}

	return it, nil
}

// TODO do something smarter with errors
func (s *subscriptionRegistry) eventHandler(sub *subscription) func([]byte) {
	index := 0
	return func(body []byte) {
		if index >= len(sub.Services) {
			index = 0
		}

		service := sub.Services[index]
		index++

		url := fmt.Sprintf("http://%s:%d%s%s", service.Address, service.Port, service.Context, service.Path)
		req, err := client.POSTBytes(url, body, "application/json")

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
			log.Printf("Call to %s:%d from %s did not return 200\n", service.Address, service.Port, sub.Topic)
		}
	}
}
