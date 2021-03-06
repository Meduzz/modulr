package modulr

import (
	"github.com/Meduzz/modulr/event"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/Meduzz/modulr/registry"
	"github.com/nats-io/nats.go"
)

// NewInMemoryServiceRegistry returns a new in memory service registry
func NewInMemoryServiceRegistry() registry.ServiceRegistry {
	return registry.NewServiceRegistry()
}

// NewHttpProxy returns a new http proxy
func NewHttpProxy(serviceRegistry registry.ServiceRegistry, factory loadbalancer.LoadBalancerFactory) registry.HttpProxy {
	return registry.NewHttpProxy(serviceRegistry, factory)
}

// NewEventProxy returns a new event proxy
func NewEventProxy(serviceRegistry registry.ServiceRegistry, factory loadbalancer.LoadBalancerFactory, eventAdapter event.EventAdapter, deliveryAdapter event.DeliveryAdapter) registry.EventProxy {
	return registry.NewEventProxy(serviceRegistry, eventAdapter, deliveryAdapter, factory)
}

// NewRoundRobinLoadBalancerFactory returns a new rr load balancer factory
func NewRoundRobinLoadBalancerFactory() loadbalancer.LoadBalancerFactory {
	return loadbalancer.NewRoundRobinFactory()
}

// NewNatsEventAdapter returns a new nats based event adapter
func NewNatsEventAdapter(conn *nats.Conn) event.EventAdapter {
	return event.NewNatsAdapter(conn)
}

// NewHttpDeliveryAdapter returns a new default delivery adapter
func NewHttpDeliveryAdapter() event.DeliveryAdapter {
	return event.NewHttpDeliveryAdapter()
}
