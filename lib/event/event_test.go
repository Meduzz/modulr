package event

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/lib/registry"
)

var (
	logg            = make(chan string, 100)
	eventadapter    = &ea{nil, true, true, true}
	deliveryadapter = &da{true}
	service         = &api.DefaultService{
		ID:      "1",
		Name:    "test",
		Address: "localhost",
		Port:    6060,
		Context: "",
		Type:    "http",
		Scheme:  "http",
		Subscriptions: []*api.Subscription{
			{
				Topic:   "test",
				Routing: "test",
				Group:   "test",
				Path:    "/webhook",
				Secret:  "top secret",
			},
		},
	}
	register     = registry.NewServiceRegistry()
	eventSupport = NewEventSupport(register)
)

type (
	ea struct {
		handler          func([]byte)
		AllowSubscribe   bool
		AllowUnsubscribe bool
		AllowPublish     bool
	}

	da struct {
		AllowDeliver bool
	}

	rf struct{}

	rr struct{}
)

// all these tests depends on each other :(

func TestMain(m *testing.M) {
	eventSupport.RegisterDeliverer("http", deliveryadapter)
	eventSupport.SetEventAdapter(eventadapter)
	eventSupport.SetLoadBalancerFactory(&rf{})

	os.Exit(m.Run())
}

func TestUnhappySubscribe(t *testing.T) {
	eventadapter.AllowSubscribe = false
	err := eventSupport.RegisterService(service)

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "subscribe" {
		t.Errorf("got the wrong error, got: %s", err.Error())
	}

	if len(logg) > 0 {
		t.Error("the log is not empty")
	}

	eventadapter.AllowSubscribe = true
}

func TestHappySubscribe(t *testing.T) {
	err := eventSupport.RegisterService(service)

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

func TestUnhappyUnsubscribe(t *testing.T) {
	eventadapter.AllowUnsubscribe = false

	err := eventSupport.DeregisterService(service)

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "unsubscribe" {
		t.Errorf("got the wrong error, got: %s", err.Error())
	}

	if len(logg) > 0 {
		t.Error("log is not empty")
	}

	eventadapter.AllowUnsubscribe = true
}

func TestHappyUnsubscribeUnhappySubscribe(t *testing.T) {
	service2 := &api.DefaultService{}
	*service2 = *service
	service2.ID = "2"
	service2.Port = 1026

	eventadapter.AllowSubscribe = false
	eventSupport.RegisterService(service2)

	err := eventSupport.DeregisterService(service)

	if err != nil {
		t.Error(err)
	}

	topic := <-logg
	if topic != "test test test" {
		t.Error("topic did not match")
	}

	if len(logg) > 0 {
		t.Error("log is not empty")
	}

	eventadapter.AllowSubscribe = true
}

func TestHappyUnsubscribeHappySubscribe(t *testing.T) {
	service2 := &api.DefaultService{}
	*service2 = *service
	service2.ID = "2"
	service2.Port = 1026

	err := eventSupport.DeregisterService(service2)

	if err != nil {
		t.Error(err)
	}

	err = eventSupport.DeregisterService(service)

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
	if !e.AllowSubscribe {
		return fmt.Errorf("subscribe")
	}

	e.handler = handler
	logg <- fmt.Sprintf("%s %s %s", topic, routing, group)
	return nil
}

func (e *ea) Unsubscribe(topic, routing, group string) error {
	if !e.AllowUnsubscribe {
		return fmt.Errorf("unsubscribe")
	}

	logg <- fmt.Sprintf("%s %s %s", topic, routing, group)

	return nil
}

func (e *ea) Publish(topic, routing string, event []byte) error {
	if !e.AllowPublish {
		return fmt.Errorf("publish")
	}

	if e.handler != nil {
		logg <- fmt.Sprintf("%s %s %s", topic, routing, string(event))
		e.handler(event)
		return nil
	}

	return fmt.Errorf("e.handler was nil")
}

func (e *ea) Request(topic, routing string, body []byte, maxWait string) ([]byte, error) {
	_, err := time.ParseDuration(maxWait)

	if err != nil {
		return nil, err
	}

	if !e.AllowPublish {
		return nil, fmt.Errorf("publish")
	}

	return body, nil
}

func (d *da) Deliver(service api.Service, sub *api.Subscription, event []byte) error {
	if !d.AllowDeliver {
		return fmt.Errorf("deliver")
	}

	logg <- fmt.Sprintf("%s %s", sub.Path, string(event))
	return nil
}

func (r *rf) For(string) api.LoadBalancer {
	return &rr{}
}

func (r *rr) Next(svcs []api.Service) api.Service {
	return svcs[0]
}
