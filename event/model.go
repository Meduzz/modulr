package event

import (
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/registry"
)

type (
	// EventDelivery - focuses on how to deliver events received
	EventDelivery interface {
		registry.Lifecycle
		RegisterDeliverer(string, Deliverer)
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

	// Deliverer - focues on delivering individual events to a certain type of service
	Deliverer interface {
		// DeliverEvent is called when there's an event to deliver. Params are service, subscription & body.
		DeliverEvent(api.Service, *api.Subscription, []byte) error
	}
)
