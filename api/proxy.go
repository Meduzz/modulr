package api

import (
	"github.com/gin-gonic/gin"
)

type (
	// Proxy - interface to forward http requests
	Proxy interface {
		// ForwarderFor - looks through internal registry for Forwarders matching the provided service
		ForwarderFor(string) (gin.HandlerFunc, error)

		// RegisterForwarder - allows us ot register forwarders for service types
		RegisterForwarder(string, Forwarder)

		// SetLoadbalancerFactory - allows us to register a loadbalancer factory
		SetLoadBalancer(LoadBalancer)
	}

	// Forwarder - interface defining the adapter that forwards the actual request and returns the actual response
	Forwarder interface {
		Handler(Service) gin.HandlerFunc
	}
)
