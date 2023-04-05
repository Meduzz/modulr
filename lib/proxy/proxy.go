package proxy

import (
	"net/http"

	"github.com/Meduzz/modulr/api"
)

type (
	proxy struct {
		registry        map[string]api.Forwarder
		serviceRegistry api.ServiceRegistry
		factory         api.LoadBalancerFactory
	}
)

// NewProxy - creates a new http loadbalancer
func NewProxy(serviceRegistry api.ServiceRegistry) api.Proxy {
	forwarders := make(map[string]api.Forwarder)

	return &proxy{
		registry:        forwarders,
		serviceRegistry: serviceRegistry,
	}
}

func (p *proxy) ForwarderFor(name string) (http.HandlerFunc, error) {
	lb := p.factory.For(name)
	services, err := p.serviceRegistry.Lookup(name)

	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		// TODO also write a pesky log about it?
		return http.NotFound, nil
	}

	service := lb.Next(services)

	forwarder, ok := p.registry[service.GetType()]

	if !ok {
		// TODO also write a pesky log about it?
		return http.NotFound, nil
	}

	return forwarder.Handler(service), nil
}

func (p *proxy) RegisterForwarder(typ string, forwarder api.Forwarder) {
	p.registry[typ] = forwarder
}

func (p *proxy) SetLoadBalancerFactory(factory api.LoadBalancerFactory) {
	p.factory = factory
}
