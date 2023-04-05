package api

type (
	// LoadBalancer - super simple interface for load balancing
	LoadBalancer interface {
		// Next - find the next service in the pool of services
		Next([]Service) Service
	}

	// LoadBalancerFactory - provides a way to create loadbalancers
	LoadBalancerFactory interface {
		// For - fetches the loadbalancer for the provided service name
		For(string) LoadBalancer
	}
)
