package algs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WeightedRoundRobinAlgorithm struct {
	Servers        []IBackendServer
	healthyServers map[uuid.UUID]IBackendServer
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
	return nil, nil
}
func (r *WeightedRoundRobinAlgorithm) healthCheck() {
	for _, server := range r.Servers {
		if err := Ping(server); err != nil {
			fmt.Printf("Server %s is unhealthy: %v\n", server.GetUrl(), err)
			server.SetStatus(UnHealthy)
			delete(r.healthyServers, server.GetID())
		} else {
			server.SetStatus(Healthy)
			if _, exists := r.healthyServers[server.GetID()]; !exists {
				r.healthyServers[server.GetID()] = server
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

	alg := &WeightedRoundRobinAlgorithm{
		Servers:        params.Servers,
		healthyServers: healthyServers,
		CurrentIndex:   -1,
		ticker:         time.NewTicker(time.Second * 30),
	}
	go func() {
		for range alg.ticker.C {
			fmt.Printf("[WeightedRoundRobinAlgorithm] health check at %v\n", time.Now())
			alg.healthCheck()
		}
	}()
	return alg, nil
}
