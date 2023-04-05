package api

import (
	"encoding/json"
)

type (
	// Service - the service details the lib need to do its job
	Service interface {
		GetID() string
		GetName() string
		GetAddress() string
		GetPort() int
		GetContext() string
		GetSubscriptions() []*Subscription
		GetScheme() string
		GetType() string
	}

	// DefaultService - implements a service
	DefaultService struct {
		ID            string          `json:"id"`                      // used in deregister
		Name          string          `json:"name"`                    // used in path (/call/<name>/...)
		Address       string          `json:"address"`                 // ip/hostname
		Port          int             `json:"port"`                    // port 1024+
		Context       string          `json:"context"`                 // used in routing
		Subscriptions []*Subscription `json:"subscriptions,omitempty"` // event subscriptions
		Scheme        string          `json:"scheme,omitempty"`        // optional scheme (if not http)
		Type          string          `json:"type"`                    // service type, as a way to decide how to deliver the payload
	}

	// Subscription - details needed for an event subscriptions
	Subscription struct {
		Topic   string `json:"topic"`             // topic/exchange
		Routing string `json:"routing,omitempty"` // routing key
		Group   string `json:"group"`             // consumer group
		Path    string `json:"path"`              // webhook path - callbacks/my.event
		Secret  string `json:"secret,omitempty"`  // webhook secret
	}

	// Event - request to publish an event on behalf of a service
	Event struct {
		Topic   string          `json:"topic"`
		Routing string          `json:"routing"`
		Body    json.RawMessage `json:"body"`
	}
)

func (s *DefaultService) GetID() string {
	return s.ID
}

func (s *DefaultService) GetName() string {
	return s.Name
}

func (s *DefaultService) GetAddress() string {
	return s.Address
}

func (s *DefaultService) GetPort() int {
	return s.Port
}

func (s *DefaultService) GetContext() string {
	return s.Context
}

func (s *DefaultService) GetSubscriptions() []*Subscription {
	return s.Subscriptions
}

func (s *DefaultService) GetScheme() string {
	return s.Scheme
}

func (s *DefaultService) GetType() string {
	return s.Type
}
