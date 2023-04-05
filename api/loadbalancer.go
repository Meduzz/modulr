package api

type (
	// LoadBalancer - super simple interface for load balancing
	LoadBalancer interface {
		// Next - find the next service in the pool of services
		Next([]Service) Service
	}
)
