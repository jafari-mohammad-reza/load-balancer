package algs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type WeightedRoundRobinAlgorithm struct {
	Servers        []IBackendServer
	healthyServers map[uuid.UUID]IBackendServer
	orderedHealthy []IBackendServer
	mu             sync.Mutex
	CurrentIndex   int
	ticker         *time.Ticker
}

func (r *WeightedRoundRobinAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *WeightedRoundRobinAlgorithm) HealthyServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *WeightedRoundRobinAlgorithm) NextServer() (IBackendServer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentIndex = (r.CurrentIndex + 1) % len(r.orderedHealthy)
	return r.orderedHealthy[r.CurrentIndex], nil
}
func (r *WeightedRoundRobinAlgorithm) healthCheck() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.healthyServers = make(map[uuid.UUID]IBackendServer)
	r.orderedHealthy = r.orderedHealthy[:0]
	for _, server := range r.Servers {
		if err := Ping(server); err != nil {
			server.SetStatus(UnHealthy)
			delete(r.healthyServers, server.GetID())
		} else {
			server.SetStatus(Healthy)
			if _, exists := r.healthyServers[server.GetID()]; !exists {
				r.healthyServers[server.GetID()] = server
			}
			weight := server.GetWeight()
			for i := 0; i < weight; i++ {
				r.orderedHealthy = append(r.orderedHealthy, server)
			}
		}
	}
}

func NewWeightedRoundRobinAlgorithm(params AlgParams) (*WeightedRoundRobinAlgorithm, error) {
	healthyServers := make(map[uuid.UUID]IBackendServer, len(params.Servers))
	for _, server := range params.Servers {
		if server.GetStatus() == Healthy {
			healthyServers[server.GetID()] = server
		}
	}
	orderedHealthy := make([]IBackendServer, 0, len(healthyServers))
	for _, server := range params.Servers {
		if server.GetStatus() == Healthy {
			healthyServers[server.GetID()] = server
			for i := 0; i < server.GetWeight(); i++ {
				orderedHealthy = append(orderedHealthy, server)
			}
		}
	}
	alg := &WeightedRoundRobinAlgorithm{
		Servers:        params.Servers,
		healthyServers: healthyServers,
		orderedHealthy: orderedHealthy,
		mu:             sync.Mutex{},
		CurrentIndex:   -1,
		ticker:         time.NewTicker(time.Second * 30),
	}
	go func() {
		for range alg.ticker.C {
			alg.healthCheck()
		}
	}()
	return alg, nil
}
