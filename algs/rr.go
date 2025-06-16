package algs

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type RoundRobinAlgorithm struct {
	Servers        []IBackendServer
	healthyServers map[uuid.UUID]IBackendServer
	orderedHealthy []IBackendServer
	CurrentIndex   int
	mu             sync.Mutex
	ticker         *time.Ticker
}

func (r *RoundRobinAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *RoundRobinAlgorithm) HealthyServers() ([]IBackendServer, error) {
	servers := make([]IBackendServer, 0, len(r.healthyServers))
	for _, server := range r.healthyServers {
		servers = append(servers, server)
	}
	return servers, nil
}
func (r *RoundRobinAlgorithm) NextServer() (IBackendServer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.orderedHealthy) == 0 {
		return nil, errors.New("no server available")
	}
	r.CurrentIndex = (r.CurrentIndex + 1) % len(r.healthyServers)
	return r.orderedHealthy[r.CurrentIndex], nil
}
func (r *RoundRobinAlgorithm) healthCheck() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.healthyServers = make(map[uuid.UUID]IBackendServer)
	r.orderedHealthy = r.orderedHealthy[:0]
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

func NewRoundRobinAlgorithm(params AlgParams) (*RoundRobinAlgorithm, error) {
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
			orderedHealthy = append(orderedHealthy, server)
		}
	}
	alg := &RoundRobinAlgorithm{
		Servers:        params.Servers,
		healthyServers: healthyServers,
		orderedHealthy: orderedHealthy,
		CurrentIndex:   -1,
		ticker:         time.NewTicker(time.Second * 30),
		mu:             sync.Mutex{},
	}
	go func() {
		for range alg.ticker.C {
			fmt.Printf("[RoundRobinAlgorithm] health check at %v\n", time.Now())
			alg.healthCheck()
		}
	}()
	return alg, nil
}
