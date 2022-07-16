package loadbalancer

import "github.com/Meduzz/modulr/api"

type (
	roundRobin struct {
		index int
	}

	roundRobinFactory struct {
		lbs map[string]LoadBalancer
	}
)

// NewRoundRobin - creates a new in memory round robin load balancer
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
	lbs := make(map[string]LoadBalancer)
	return &roundRobinFactory{lbs}
}

func (f *roundRobinFactory) For(name string) LoadBalancer {
	lb, exists := f.lbs[name]

	if !exists {
		lb = NewRoundRobin()
		f.lbs[name] = lb
	}

	return lb
}
