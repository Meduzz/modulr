package proxy

import (
	"net/http"

	"github.com/Meduzz/modulr/registry"
)

type (
	// LoadBalancer - interface for loadbalancing
	LoadBalancer interface {
		registry.Lifecycle
		// Lookup - returns a http.HandlerFunc or nil
		Lookup(string) http.HandlerFunc
	}
)
