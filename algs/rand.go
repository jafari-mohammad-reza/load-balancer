package algs

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

type RandomAlgorithm struct {
	Servers        []IBackendServer
	healthyServers map[uuid.UUID]IBackendServer
	ticker         *time.Ticker
}

func (r *RandomAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *RandomAlgorithm) HealthyServers() ([]IBackendServer, error) {
	servers := make([]IBackendServer, 0, len(r.healthyServers))
	for _, server := range r.healthyServers {
		servers = append(servers, server)
	}
	return servers, nil
}
func (r *RandomAlgorithm) NextServer() (IBackendServer, error) {
	if len(r.Servers) == 0 || len(r.healthyServers) == 0 {
		return nil, errors.New("no servers available")
	}
	randIndex := rand.IntN(len(r.healthyServers))
	healthyServers := make([]IBackendServer, 0, len(r.healthyServers))
	for _, server := range r.healthyServers {
		healthyServers = append(healthyServers, server)
	}
	return healthyServers[randIndex], nil
}
func (r *RandomAlgorithm) healthCheck() {
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

func NewRandomAlgorithm(params AlgParams) (*RandomAlgorithm, error) {
	healthyServers := make(map[uuid.UUID]IBackendServer, len(params.Servers))
	for _, server := range params.Servers {
		if server.GetStatus() == Healthy {
			healthyServers[server.GetID()] = server
		}
	}
	alg := &RandomAlgorithm{
		Servers:        params.Servers,
		healthyServers: healthyServers,
		ticker:         time.NewTicker(time.Second * 1),
	}
	go func() {
		for range alg.ticker.C {
			fmt.Printf("[RandomAlgorithm] health check at %v\n", time.Now())
			alg.healthCheck()
		}
	}()
	return alg, nil
}
