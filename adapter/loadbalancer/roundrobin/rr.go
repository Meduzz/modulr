package roundrobin

import (
	"github.com/Meduzz/modulr"
	"github.com/Meduzz/modulr/api"
)

type (
	roundRobin struct {
		index map[string]int
	}
)

func init() {
	rr := NewRoundRobin()

	modulr.EventSupport.SetLoadBalancer(rr)
	modulr.HttpProxy.SetLoadBalancer(rr)
}

// NewRoundRobin - creates a new in memory round robin load balancer
func NewRoundRobin() api.LoadBalancer {
	idx := make(map[string]int)
	return &roundRobin{idx}
}

func (r *roundRobin) Next(pool []api.Service) api.Service {
	if len(pool) == 0 {
		return nil
	}

	index, ok := r.index[pool[0].GetName()]

	if !ok {
		index = -1
	}

	index += 1

	if index >= len(pool) {
		index = 0
	}

	winner := pool[index]

	r.index[winner.GetName()] = index

	return winner
}
