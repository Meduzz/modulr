package roundrobin

import (
	"sync/atomic"

	"github.com/Meduzz/modulr"
	"github.com/Meduzz/modulr/api"
)

type (
	roundRobin struct {
		index map[string]*atomic.Int32
	}
)

var LoadBalancer api.LoadBalancer

func init() {
	LoadBalancer := NewRoundRobin()

	modulr.EventSupport.SetLoadBalancer(LoadBalancer)
	modulr.HttpProxy.SetLoadBalancer(LoadBalancer)
}

// NewRoundRobin - creates a new in memory round robin load balancer
func NewRoundRobin() api.LoadBalancer {
	idx := make(map[string]*atomic.Int32)
	return &roundRobin{idx}
}

func (r *roundRobin) Next(pool []api.Service) api.Service {
	if len(pool) == 0 {
		return nil
	}

	index, ok := r.index[pool[0].GetName()]

	if !ok {
		index.Store(-1)
	}

	index.Add(1)

	if int(index.Load()) >= len(pool) {
		index.Store(0)
	}

	winner := pool[index.Load()]

	return winner
}
