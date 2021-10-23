package registry

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/vulcand/oxy/forward"
)

type (
	// LoadBalancer - interface for loadbalancing
	LoadBalancer interface {
		Lifecycle
		// Lookup - find service by name and return a http.HandlerFunc or nil
		Lookup(string) http.HandlerFunc
	}

	httpproxy struct {
		forwarder map[string]*forward.Forwarder        // id -> forwarder
		lbs       map[string]loadbalancer.LoadBalancer // name -> loadbalancer
		registry  ServiceRegistry
		factory   loadbalancer.LoadBalancerFactory
	}

	rewriter struct {
		service api.Service
	}

	chained struct {
		rewriters []forward.ReqRewriter
	}
)

// NewLoadBalancer - creates a new http loadbalancer
func NewLoadBalancer(factory loadbalancer.LoadBalancerFactory) LoadBalancer {
	forwarders := make(map[string]*forward.Forwarder)
	lbs := make(map[string]loadbalancer.LoadBalancer)

	return &httpproxy{
		forwarder: forwarders,
		lbs:       lbs,
		factory:   factory,
	}
}

func (p *httpproxy) RegisterService(name string, service api.Service) error {
	p.lbs[name] = p.factory.Create()

	return nil
}

func (p *httpproxy) DeregisterService(name string, service api.Service) error {
	delete(p.lbs, name)

	return nil
}

func (p *httpproxy) RegisterInstance(service api.Service) error {
	// TODO errorhandling
	// TODO circuitbreaker?
	// TODO retries?
	fwd, err := forward.New(forward.Rewriter(chainedRewriters(&rewriter{service})))

	if err != nil {
		return err
	}

	p.forwarder[service.GetID()] = fwd

	log.Printf("Created loadbalanser for %s\n", service.GetName())

	return nil
}

func (p *httpproxy) DeregisterInstance(service api.Service) error {
	delete(p.forwarder, service.GetID())

	log.Printf("Removing %s from loadbalancer (%s)\n", service.GetID(), service.GetName())

	return nil
}

func (p *httpproxy) Lookup(name string) http.HandlerFunc {
	lb, exists := p.lbs[name]

	if !exists {
		// TOOD we're out of sync
		return nil
	}

	services := p.registry.Lookup(name)
	service := lb.Next(services)

	if service == nil {
		return nil
	}

	handler, exists := p.forwarder[service.GetID()]

	if !exists {
		// TODO we're out of sync...
		return nil
	}

	return handler.ServeHTTP
}

func (p *httpproxy) ServiceRegistry(registry ServiceRegistry) {
	p.registry = registry
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
