package loadbalancer

import "github.com/Meduzz/modulr/api"

type (
	// LoadBalancer - super simple interface for load balancing
	LoadBalancer interface {
		// Next - find the next service in the pool of services
		Next([]api.Service) api.Service
	}

	// LoadBalancerFactory - provides a way to create loadbalancers
	LoadBalancerFactory interface {
		// Create - create a new loadbalancer
		Create() LoadBalancer
	}
)
