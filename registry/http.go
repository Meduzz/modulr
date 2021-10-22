package registry

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

type (
	// LoadBalancer - interface for loadbalancing
	LoadBalancer interface {
		Lifecycle
		// Lookup - returns a http.HandlerFunc or nil
		Lookup(string) http.HandlerFunc
	}

	httpproxy struct {
		lbs map[string]*roundrobin.RoundRobin // name -> loadbalancer
	}

	rewriter struct {
		name string
	}

	chained struct {
		rewriters []forward.ReqRewriter
	}
)

// NewLoadBalancer - creates a new http loadbalancer
func NewLoadBalancer() LoadBalancer {
	lbs := make(map[string]*roundrobin.RoundRobin)

	return &httpproxy{
		lbs: lbs,
	}
}

func (p *httpproxy) Register(service api.Service) error {
	lb, exists := p.lbs[service.GetName()]

	if !exists {
		// TODO errorhandling
		// TODO circuitbreaker?
		// TODO retries?
		fwd, _ := forward.New(forward.Rewriter(chainedRewriters(&rewriter{service.GetName()})))
		rr, _ := roundrobin.New(fwd)
		p.lbs[service.GetName()] = rr
		lb = rr

		log.Printf("Created loadbalanser for %s\n", service.GetName())
	}

	serviceUrl := service.ToURL()

	log.Printf("Adding %s to loadbalancer (%s)\n", serviceUrl.String(), service.GetName())

	return lb.UpsertServer(serviceUrl)
}

func (p *httpproxy) Deregister(service api.Service) error {
	lb, exists := p.lbs[service.GetName()]

	if !exists {
		return nil
	}

	serviceUrl := service.ToURL()

	log.Printf("Removing %s from loadbalancer (%s)\n", serviceUrl.String(), service.GetName())

	err := lb.RemoveServer(serviceUrl)

	if err != nil {
		return err
	}

	if len(lb.Servers()) == 0 {
		delete(p.lbs, service.GetName())
	}

	return nil
}

func (p *httpproxy) Lookup(name string) http.HandlerFunc {
	it, exists := p.lbs[name]

	if !exists {
		return nil
	}

	return it.ServeHTTP
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

// Request rewriter.
func (r *rewriter) Rewrite(req *http.Request) {
	req.URL.RawPath = strings.Replace(req.URL.RawPath, fmt.Sprintf("/call/%s", r.name), "", 1)
	req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/call/%s", r.name), "", 1)
}

func (c *chained) Rewrite(req *http.Request) {
	for _, r := range c.rewriters {
		r.Rewrite(req)
	}
}
