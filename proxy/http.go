package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/gin-gonic/gin"
	"github.com/vulcand/oxy/forward"
)

type (
	// HttpProxy - interface for http loadbalancing
	HttpProxy interface {
		// ForwarderFor - find service by name and return a gin.HandlerFunc or nil
		ForwarderFor(api.Service) gin.HandlerFunc
	}

	httpproxy struct {
		factory loadbalancer.LoadBalancerFactory
	}

	rewriter struct {
		service api.Service
	}

	chained struct {
		rewriters []forward.ReqRewriter
	}
)

// NewHttpProxy - creates a new http loadbalancer
func NewHttpProxy(factory loadbalancer.LoadBalancerFactory) HttpProxy {
	return &httpproxy{
		factory: factory,
	}
}

func (p *httpproxy) ForwarderFor(service api.Service) gin.HandlerFunc {
	// TODO errorhandling
	// TODO circuitbreaker?
	// TODO retries?
	handler, err := forward.New(forward.Rewriter(chainedRewriters(&rewriter{service})))

	if err != nil {
		return nil
	}

	return gin.WrapF(handler.ServeHTTP)
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
	req.URL.RawPath = strings.Replace(req.URL.RawPath, fmt.Sprintf("/call/%s", r.service.GetName()), r.service.GetContext(), 1)
	req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/call/%s", r.service.GetName()), r.service.GetContext(), 1)
	req.URL.Scheme = r.service.GetScheme()

	if r.service.GetPort() != 0 {
		req.URL.Host = fmt.Sprintf("%s:%d", r.service.GetAddress(), r.service.GetPort())
	} else {
		req.URL.Host = r.service.GetAddress()
	}
}

// Chained request rewriter
func (c *chained) Rewrite(req *http.Request) {
	for _, r := range c.rewriters {
		r.Rewrite(req)
	}
}
