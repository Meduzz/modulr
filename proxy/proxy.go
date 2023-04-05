package proxy

import (
	"net/http"

	"github.com/Meduzz/modulr/api"
)

type (
	proxy struct {
		registry map[string]Forwarder
	}
)

// NewProxy - creates a new http loadbalancer
func NewProxy() Proxy {
	forwarders := make(map[string]Forwarder)
	return &proxy{forwarders}
}

func (p *proxy) ForwarderFor(service api.Service) http.HandlerFunc {
	forwarder, ok := p.registry[service.GetType()]

	if !ok {
		// TODO also write a pesky log about it?
		return http.NotFound
	}

	return forwarder.Handler(service)
}

func (p *proxy) RegisterForwarder(typ string, forwarder Forwarder) {
	p.registry[typ] = forwarder
}
