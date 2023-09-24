package roundrobin

import (
	"sync"

	"github.com/Meduzz/modulr"
	"github.com/Meduzz/modulr/api"
)

type (
	roundRobin struct {
		index map[string]int
		lock  *sync.Mutex
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
	idx := make(map[string]int)
	return &roundRobin{idx, &sync.Mutex{}}
}

func (r *roundRobin) Next(pool []api.Service) api.Service {
	if len(pool) == 0 {
		return nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	index, ok := r.index[pool[0].GetName()]

	if !ok {
		index = -1
	}

	index = index + 1

	if index >= len(pool) {
		index = 0
	}

	r.index[pool[0].GetName()] = index

	winner := pool[index]

	return winner
}
