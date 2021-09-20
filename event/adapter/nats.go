package adapter

import (
	"fmt"

	"github.com/Meduzz/modulr/event"
	"github.com/nats-io/nats.go"
)

type (
	adapter struct {
		conn *nats.Conn
		subs map[string]*nats.Subscription
	}
)

func NewNatsAdapter(conn *nats.Conn) event.EventAdapter {
	subs := make(map[string]*nats.Subscription)

	return &adapter{
		conn: conn,
		subs: subs,
	}
}

func (a *adapter) Publish(topic string, routing string, body []byte) error {
	return a.conn.Publish(topic, body)
}

func (a *adapter) Subscribe(topic, routing, group string, handler func([]byte)) error {
	key := fmt.Sprintf("%s.%s.%s", topic, routing, group)

	if group != "" {
		sub, err := a.conn.QueueSubscribe(topic, group, func(msg *nats.Msg) {
			handler(msg.Data)
		})

		if err != nil {
			return err
		}

		a.subs[key] = sub
	} else {
		sub, err := a.conn.Subscribe(topic, func(msg *nats.Msg) {
			handler(msg.Data)
		})

		if err != nil {
			return err
		}

		a.subs[key] = sub
	}

	return nil
}

func (a *adapter) Unsubscribe(topic, routing, group string) error {
	key := fmt.Sprintf("%s.%s.%s", topic, routing, group)
	sub, exists := a.subs[key]

	if exists {
		return sub.Unsubscribe()
	}

	return nil
}
