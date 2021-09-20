package httpadapter

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/registry"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

type (
	// LoadBalancer - interface for loadbalancing
	LoadBalancer interface {
		registry.Lifecycle
		// Lookup - returns a http.HandlerFunc or nil
		Lookup(string) http.HandlerFunc
	}

	proxy struct {
		lbs map[string]*roundrobin.RoundRobin // name -> loadbalancer
	}

	rewriter struct {
		name string
	}
)

// NewLoadBalancer - creates a new http loadbalancer
func NewLoadBalancer() LoadBalancer {
	lbs := make(map[string]*roundrobin.RoundRobin)

	return &proxy{
		lbs: lbs,
	}
}

func (p *proxy) Register(service *api.Service) error {
	lb, exists := p.lbs[service.Name]

	if !exists {
		// TODO errorhandling
		// TODO circuitbreaker?
		// TODO retries?
		fwd, _ := forward.New(forward.Rewriter(&rewriter{service.Name}))
		rr, _ := roundrobin.New(fwd)
		p.lbs[service.Name] = rr
		lb = rr

		log.Printf("Created loadbalanser for %s\n", service.Name)
	}

	serviceUrl := &url.URL{
		Scheme:  "http",
		Host:    fmt.Sprintf("%s:%d", service.Address, service.Port),
		Path:    service.Context,
		RawPath: service.Context,
	}

	log.Printf("Adding %s to loadbalancer (%s)\n", serviceUrl.String(), service.Name)

	return lb.UpsertServer(serviceUrl)
}

func (p *proxy) Deregister(service *api.Service) error {
	lb, exists := p.lbs[service.Name]

	if !exists {
		return nil
	}

	serviceUrl := &url.URL{
		Scheme:  "http",
		Host:    fmt.Sprintf("%s:%d", service.Address, service.Port),
		Path:    service.Context,
		RawPath: service.Context,
	}

	log.Printf("Removing %s from loadbalancer (%s)\n", serviceUrl.String(), service.Name)

	err := lb.RemoveServer(serviceUrl)

	if err != nil {
		return err
	}

	if len(lb.Servers()) == 0 {
		delete(p.lbs, service.Name)
	}

	return nil
}

func (p *proxy) Lookup(name string) http.HandlerFunc {
	it, exists := p.lbs[name]

	if !exists {
		return nil
	}

	return it.ServeHTTP
}

func (r *rewriter) Rewrite(req *http.Request) {
	req.URL.RawPath = strings.Replace(req.URL.RawPath, fmt.Sprintf("/call/%s", r.name), "", 1)
	req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/call/%s", r.name), "", 1)
}
