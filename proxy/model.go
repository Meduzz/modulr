package proxy

import (
	"net/http"

	"github.com/Meduzz/modulr/api"
)

type (
	// Proxy - interface to forward http requests
	Proxy interface {
		// ForwarderFor - looks through internal registry for Forwarders matching the provided service
		ForwarderFor(api.Service) http.HandlerFunc

		// RegisterForwarder - allows us ot register forwarders for service types
		RegisterForwarder(string, Forwarder)
	}

	// Forwarder - interface defining the adapter that forwards the actual request and returns the actual response
	Forwarder interface {
		Handler(api.Service) http.HandlerFunc
	}
)
