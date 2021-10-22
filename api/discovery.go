package api

import (
	"encoding/json"
	"fmt"
	"net/url"
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
		ToURL() *url.URL
	}

	// DefaultService - implements a service
	DefaultService struct {
		ID            string          `json:"id"`                      // used in deregister
		Name          string          `json:"name"`                    // used in path (/call/<name>/...)
		Address       string          `json:"address"`                 // ip/hostname
		Port          int             `json:"port"`                    // port 1024+
		Context       string          `json:"context"`                 // used in routing
		Subscriptions []*Subscription `json:"subscriptions,omitempty"` // event subscriptions
	}

	// Subscription - details needed for an event subscriptions
	Subscription struct {
		Topic   string `json:"topic"`   // topic/exchange
		Routing string `json:"routing"` // routing key
		Group   string `json:"group"`   // consumer group
		Path    string `json:"path"`    // webhook path - service.context.
	}

	// Event - request to publish an event on behalf of a service
	Event struct {
		Topic   string          `json:"topic"`
		Routing string          `json:"routing"`
		Body    json.RawMessage `json:"body"`
	}
)

func (s *DefaultService) ToURL() *url.URL {
	// TODO only a matter of time until https
	return &url.URL{
		Scheme:  "http",
		Host:    fmt.Sprintf("%s:%d", s.Address, s.Port),
		Path:    s.Context,
		RawPath: s.Context,
	}
}

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
