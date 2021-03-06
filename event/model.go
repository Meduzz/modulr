package event

type (
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

	DeliveryAdapter interface {
		DeliverEvent(string, []byte) error
	}
)
