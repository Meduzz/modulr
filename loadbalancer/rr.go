package loadbalancer

import "github.com/Meduzz/modulr/api"

type (
	roundRobin struct {
		index int
	}

	roundRobinFactory struct{}
)

// NewRoundRobin - creates a new round robin load balancer
func NewRoundRobin() LoadBalancer {
	return &roundRobin{-1}
}

func (r *roundRobin) Next(pool []api.Service) api.Service {
	r.index = r.index + 1

	if len(pool) == 0 {
		return nil
	}

	if r.index >= len(pool) {
		r.index = 0
	}

	return pool[r.index]
}

func NewRoundRobinFactory() LoadBalancerFactory {
	return &roundRobinFactory{}
}

func (f *roundRobinFactory) Create() LoadBalancer {
	return NewRoundRobin()
}
