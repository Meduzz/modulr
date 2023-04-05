package event

import (
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/registry"
)

type (
	// EventSupport - focuses on how to deliver events received
	EventSupport interface {
		registry.Lifecycle
		RegisterDeliverer(string, EventDeliveryAdapter)
		SetEventAdapter(EventAdapter)
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
		Deliver(api.Service, *api.Subscription, []byte) error
	}
)
