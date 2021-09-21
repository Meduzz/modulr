package adapter

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/proxy"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

type (
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
func NewLoadBalancer() proxy.LoadBalancer {
	lbs := make(map[string]*roundrobin.RoundRobin)

	return &httpproxy{
		lbs: lbs,
	}
}

func (p *httpproxy) Register(service *api.Service) error {
	lb, exists := p.lbs[service.Name]

	if !exists {
		// TODO errorhandling
		// TODO circuitbreaker?
		// TODO retries?
		fwd, _ := forward.New(forward.Rewriter(chainedRewriters(&rewriter{service.Name})))
		rr, _ := roundrobin.New(fwd)
		p.lbs[service.Name] = rr
		lb = rr

		log.Printf("Created loadbalanser for %s\n", service.Name)
	}

	serviceUrl := service.ToURL()

	log.Printf("Adding %s to loadbalancer (%s)\n", serviceUrl.String(), service.Name)

	return lb.UpsertServer(serviceUrl)
}

func (p *httpproxy) Deregister(service *api.Service) error {
	lb, exists := p.lbs[service.Name]

	if !exists {
		return nil
	}

	serviceUrl := service.ToURL()

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
	list = append(list, &forward.HeaderRewriter{false, ""})

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
