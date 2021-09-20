package delivery

type (
	DeliveryAdapter interface {
		DeliverEvent(string, []byte) error
	}
)
