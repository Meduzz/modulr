package roundrobin

import (
	"fmt"
	"testing"
	"time"

	"github.com/Meduzz/modulr/api"
)

func TestRoundRobin(t *testing.T) {
	subject := NewRoundRobin()

	emptyPool := make([]api.Service, 0)
	ordinaryPool := make([]api.Service, 0)

	ordinaryPool = append(ordinaryPool, &api.DefaultService{ID: "1"})
	ordinaryPool = append(ordinaryPool, &api.DefaultService{ID: "2"})

	t.Run("with empty pool", func(t *testing.T) {
		result := subject.Next(emptyPool)

		if result != nil {
			t.Error("result was not nil")
		}
	})

	t.Run("with ordinary pool", func(t *testing.T) {
		// get first node
		result := subject.Next(ordinaryPool)

		if result == nil {
			t.Error("result was nil")
		}

		if result.GetID() != "1" {
			t.Errorf("result id was not 1 but %s", result.GetID())
		}

		// iterate to node 2
		result = subject.Next(ordinaryPool)

		if result == nil {
			t.Error("result was nil")
		}

		if result.GetID() != "2" {
			t.Errorf("result id was not 2 but %s", result.GetID())
		}

		// back at first node
		result = subject.Next(ordinaryPool)

		if result == nil {
			t.Error("result was nil")
		}

		if result.GetID() != "1" {
			t.Errorf("result id was not 1 but %s", result.GetID())
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		results := make(chan api.Service, 1000)
		start := time.Now()
		go spawn(subject, ordinaryPool, 1000, results)
		go spawn(subject, ordinaryPool, 1000, results)
		go spawn(subject, ordinaryPool, 1000, results)

		count := 0
		for range results {
			count++

			if count == 3000 {
				close(results)
			}
		}

		end := time.Now()

		fmt.Printf("Done in %s\n", end.Sub(start).String())
	})
}

func spawn(subject api.LoadBalancer, pool []api.Service, count int, callback chan api.Service) {
	for i := 0; i < count; i++ {
		callback <- subject.Next(pool)
	}
}
