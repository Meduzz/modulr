package registry

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Meduzz/modulr/api"
)

var (
	logg            = make(chan string, 100)
	eventadapter    = &ea{nil}
	deliveryadapter = &da{}
	subscribe       = false
	unsubscribe     = false
	publish         = false
	deliver         = false
	subject         = NewEventRegistry(eventadapter, deliveryadapter)
	service         = &api.DefaultService{
		ID:      "1",
		Name:    "test",
		Address: "localhost",
		Port:    1025,
		Context: "/test",
		Subscriptions: []*api.Subscription{
			&api.Subscription{
				Topic:   "test",
				Routing: "test",
				Group:   "test",
				Path:    "/event",
			},
		},
	}
	testEvent = &api.Event{
		Topic:   "test",
		Routing: "test",
		Body:    json.RawMessage([]byte("test")),
	}
)

type (
	ea struct {
		handler func([]byte)
	}

	da struct {
	}
)

// all these tests depends on each other :-(

func TestUnhappySubscribe(t *testing.T) {
	subscribe = true
	err := subject.Register(service)

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "subscribe" {
		t.Errorf("got the wrong error, got: %s", err.Error())
	}

	if len(logg) > 0 {
		t.Error("the log is not empty")
	}

	subscribe = false
}

func TestHappySubscribe(t *testing.T) {
	err := subject.Register(service)

	if err != nil {
		t.Error(err)
	}

	topic := <-logg
	if topic != "test test test" {
		t.Error("topic did not match")
	}

	if len(logg) > 0 {
		t.Error("there are more logs")
	}
}

func TestRegisterAgain(t *testing.T) {
	err := subject.Register(service)

	if err != nil {
		t.Error(err)
	}

	if len(logg) > 0 {
		t.Error("there are more logs")
	}
}

func TestHappyPublishAndDelivery(t *testing.T) {
	err := subject.Publish(testEvent)

	if err != nil {
		t.Error(err)
	}

	publish := <-logg
	if publish != "test test test" {
		t.Error("the published event did not match")
	}

	delivered := <-logg
	if delivered != "http://localhost:1025/test/event test" {
		t.Error("the delivered data did not match")
	}

	if len(logg) > 0 {
		t.Error("the log is not empty")
	}
}

func TestHappyPublishUnhappyDelivery(t *testing.T) {
	deliver = true
	err := subject.Publish(testEvent)

	if err != nil {
		t.Error(err)
	}

	publish := <-logg
	if publish != "test test test" {
		t.Error("the published event did not match")
	}

	if len(logg) > 0 {
		t.Error("the log is not empty")
	}
}

func TestUnhappyPublish(t *testing.T) {
	publish = true
	err := subject.Publish(testEvent)

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "publish" {
		t.Errorf("got the wrong error, got: %s", err.Error())
	}

	if len(logg) > 0 {
		t.Error("the log is not empty")
	}
}

func TestUnhappyUnsubscribe(t *testing.T) {
	unsubscribe = true

	err := subject.Deregister(service)

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "unsubscribe" {
		t.Errorf("got the wrong error, got: %s", err.Error())
	}

	if len(logg) > 0 {
		t.Error("log is not empty")
	}

	unsubscribe = false
}

func TestHappyUnsubscribeUnhappySubscribe(t *testing.T) {
	service2 := &api.DefaultService{}
	*service2 = *service
	service2.ID = "2"
	service2.Port = 1026

	subject.Register(service2)

	subscribe = true
	err := subject.Deregister(service)

	if err != nil {
		t.Error(err)
	}

	if len(logg) > 0 {
		t.Error("log is not empty")
	}

	subscribe = false
}

func TestHappyUnsubscribeHappySubscribe(t *testing.T) {
	service2 := &api.DefaultService{}
	*service2 = *service
	service2.ID = "2"
	service2.Port = 1026

	err := subject.Deregister(service2)

	if err != nil {
		t.Error(err)
	}

	err = subject.Deregister(service)

	if err != nil {
		t.Error(err)
	}

	topic := <-logg
	if topic != "test test test" {
		t.Errorf("topic did not match")
	}

	topic = <-logg
	if topic != "test test test" {
		t.Errorf("topic did not match")
	}

	if len(logg) > 0 {
		t.Error("log is not empty")
	}
}

func (e *ea) Subscribe(topic, routing, group string, handler func([]byte)) error {
	if subscribe {
		return fmt.Errorf("subscribe")
	}

	e.handler = handler
	logg <- fmt.Sprintf("%s %s %s", topic, routing, group)
	return nil
}

func (e *ea) Unsubscribe(topic, routing, group string) error {
	if unsubscribe {
		return fmt.Errorf("unsubscribe")
	}

	logg <- fmt.Sprintf("%s %s %s", topic, routing, group)

	return nil
}

func (e *ea) Publish(topic, routing string, event []byte) error {
	if publish {
		return fmt.Errorf("publish")
	}

	if e.handler != nil {
		logg <- fmt.Sprintf("%s %s %s", topic, routing, string(event))
		e.handler(event)
		return nil
	}

	return fmt.Errorf("e.handler was nil")
}

func (d *da) DeliverEvent(url string, event []byte) error {
	if deliver {
		return fmt.Errorf("deliver")
	}

	logg <- fmt.Sprintf("%s %s", url, string(event))
	return nil
}
