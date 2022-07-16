package registry

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/vulcand/oxy/forward"
)

type (
	// HttpProxy - interface for http loadbalancing
	HttpProxy interface {
		// Lookup - find service by name and return a http.HandlerFunc or nil
		Lookup(string) http.HandlerFunc
	}

	httpproxy struct {
		registry ServiceRegistry
		factory  loadbalancer.LoadBalancerFactory
	}

	rewriter struct {
		service api.Service
	}

	chained struct {
		rewriters []forward.ReqRewriter
	}
)

// NewHttpProxy - creates a new http loadbalancer
func NewHttpProxy(registry ServiceRegistry, factory loadbalancer.LoadBalancerFactory) HttpProxy {
	return &httpproxy{
		factory:  factory,
		registry: registry,
	}
}

func (p *httpproxy) Lookup(name string) http.HandlerFunc {
	services := p.registry.Lookup(name)
	lb := p.factory.For(name)

	service := lb.Next(services)

	if service == nil {
		return nil
	}

	// TODO errorhandling
	// TODO circuitbreaker?
	// TODO retries?
	handler, err := forward.New(forward.Rewriter(chainedRewriters(&rewriter{service})))

	if err != nil {
		return nil
	}

	return handler.ServeHTTP
}

func chainedRewriters(rewriter forward.ReqRewriter) forward.ReqRewriter {
	list := make([]forward.ReqRewriter, 0)
	list = append(list, rewriter)
	list = append(list, &forward.HeaderRewriter{
		TrustForwardHeader: false,
		Hostname:           ""})

	return &chained{
		rewriters: list,
	}
}

// Host/Path request rewriter.
func (r *rewriter) Rewrite(req *http.Request) {
	req.URL.RawPath = strings.Replace(req.URL.RawPath, fmt.Sprintf("/call/%s", r.service.GetName()), "", 1)
	req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/call/%s", r.service.GetName()), "", 1)
	req.URL.Host = fmt.Sprintf("%s:%d", r.service.GetAddress(), r.service.GetPort())
	// TODO assuming http
	req.URL.Scheme = "http"
}

// Chained request rewriter
func (c *chained) Rewrite(req *http.Request) {
	for _, r := range c.rewriters {
		r.Rewrite(req)
	}
}
