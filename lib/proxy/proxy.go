package proxy

import (
	"net/http"

	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

type (
	proxy struct {
		registry        map[string]api.Forwarder
		serviceRegistry api.ServiceRegistry
		lb              api.LoadBalancer
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

func (p *proxy) ForwarderFor(name string) (gin.HandlerFunc, error) {
	services, err := p.serviceRegistry.Lookup(name)

	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		// TODO also write a pesky log about it?
		return gin.WrapF(http.NotFound), nil
	}

	service := p.lb.Next(services)

	forwarder, ok := p.registry[service.GetType()]

	if !ok {
		// TODO also write a pesky log about it?
		return gin.WrapF(http.NotFound), nil
	}

	return forwarder.Handler(service), nil
}

func (p *proxy) RegisterForwarder(typ string, forwarder api.Forwarder) {
	p.registry[typ] = forwarder
}

func (p *proxy) SetLoadBalancer(lb api.LoadBalancer) {
	p.lb = lb
}
