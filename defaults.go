package modulr

import (
	"github.com/Meduzz/modulr/event"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/Meduzz/modulr/proxy"
	"github.com/Meduzz/modulr/registry"
	"github.com/nats-io/nats.go"
)

// NewServiceRegistry returns a new in memory service registry
func NewServiceRegistry() registry.ServiceRegistry {
	return registry.NewServiceRegistry()
}

func NewInMemoryRegistryStorage() registry.RegistryStorage {
	return registry.NewInMemoryStorage()
}

// NewHttpProxy returns a new http proxy
func NewHttpProxy() proxy.Proxy {
	return proxy.NewProxy()
}

// NewEventSupport returns a new event support
func NewEventSupport(registry registry.ServiceRegistry, factory loadbalancer.LoadBalancerFactory, eventAdapter event.EventAdapter) registry.Lifecycle {
	return event.NewEventSupport(registry, eventAdapter, factory)
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
func NewHttpDeliveryAdapter() event.Deliverer {
	return event.NewHttpDeliverer()
}
