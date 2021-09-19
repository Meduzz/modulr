package natsadapter

import (
	"fmt"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/modulr/event"
	"github.com/nats-io/nats.go"
)

type (
	adapter struct {
		conn *nats.Conn
		subs map[string]*nats.Subscription
	}
)

func NewNatsAdapter() event.EventAdapter {
	conn, err := nuts.Connect()

	if err != nil {
		// TODO
		panic(err)
	}

	subs := make(map[string]*nats.Subscription)

	return &adapter{
		conn: conn,
		subs: subs,
	}
}

func (a *adapter) Publish(topic string, routing string, body []byte) {
	a.conn.Publish(topic, body)
}

func (a *adapter) Subscribe(topic, routing, group string, handler func([]byte)) {
	key := fmt.Sprintf("%s.%s.%s", topic, routing, group)

	if group != "" {
		sub, _ := a.conn.QueueSubscribe(topic, group, func(msg *nats.Msg) {
			handler(msg.Data)
		})

		a.subs[key] = sub
	} else {
		sub, _ := a.conn.Subscribe(topic, func(msg *nats.Msg) {
			handler(msg.Data)
		})

		a.subs[key] = sub
	}
}

func (a *adapter) Unsubscribe(topic, routing, group string) {
	key := fmt.Sprintf("%s.%s.%s", topic, routing, group)
	sub, exists := a.subs[key]

	if exists {
		sub.Unsubscribe()
	}
}
