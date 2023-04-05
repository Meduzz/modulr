package roundrobin

import (
	"github.com/Meduzz/modulr"
	"github.com/Meduzz/modulr/api"
)

type (
	roundRobin struct {
		index int
	}

	roundRobinFactory struct {
		lbs map[string]api.LoadBalancer
	}
)

func init() {
	rrf := NewRoundRobinFactory()

	modulr.EventSupport.SetLoadBalancerFactory(rrf)
	modulr.HttpProxy.SetLoadBalancerFactory(rrf)
}

// NewRoundRobin - creates a new in memory round robin load balancer
func NewRoundRobin() api.LoadBalancer {
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

func NewRoundRobinFactory() api.LoadBalancerFactory {
	lbs := make(map[string]api.LoadBalancer)
	return &roundRobinFactory{lbs}
}

func (f *roundRobinFactory) For(name string) api.LoadBalancer {
	lb, exists := f.lbs[name]

	if !exists {
		lb = NewRoundRobin()
		f.lbs[name] = lb
	}

	return lb
}
