package api

import (
	"encoding/json"
)

type (
	// EventSupport - focuses on how to deliver events received
	EventSupport interface {
		Lifecycle
		Publish(*Event) error
		Request(*Event, string) ([]byte, error)
		RegisterDeliverer(string, EventDeliveryAdapter)
		SetEventAdapter(EventAdapter)
		SetLoadBalancer(LoadBalancer)
	}

	// EventAdapter - interface to be implemented by event adapters
	EventAdapter interface {
		// Publish - publish a payload to a topic with optional routingkey
		Publish(string, string, []byte) error
		// Subscribe - subscribe to a topic with optional routingkey and subscribergroup
		Subscribe(string, string, string, func([]byte)) error
		// Unsubscribe - unsubscribe to a topic with optional routingkey and subscribergroup
		Unsubscribe(string, string, string) error
		// Request - do a rpc request over the event adapter
		Request(string, string, []byte, string) ([]byte, error)
	}

	// EventDeliverer - focues on delivering individual events to a certain type of service
	EventDeliveryAdapter interface {
		// Deliver is called when there's an event to deliver. Params are service, subscription & body.
		Deliver(Service, *Subscription, []byte) error
	}

	// Event - request to publish an event on behalf of a service
	Event struct {
		Topic   string          `json:"topic"`
		Routing string          `json:"routing"`
		Body    json.RawMessage `json:"body"`
	}
)
