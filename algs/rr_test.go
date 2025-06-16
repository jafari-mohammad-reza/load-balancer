package algs

import (
	"os"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	os.Setenv("RUN_TYPE", "test")
	params := AlgParams{
		Servers: []IBackendServer{
			NewBackendServer("localhost", 8080, 1),
			NewBackendServer("localhost", 8081, 1),
			NewBackendServer("localhost", 8082, 1),
			NewBackendServer("localhost", 8083, 1),
			NewBackendServer("localhost", 8084, 1),
			NewBackendServer("localhost", 8085, 1),
		},
	}
	alg, err := NewRoundRobinAlgorithm(params)
	if err != nil {
		t.Fatalf("Failed to create RoundRobinAlgorithm: %v", err)
	}
	allServers, err := alg.AllServers()
	if err != nil || len(allServers) != len(params.Servers) {
		t.Errorf("AllServers returned unexpected result: %v, error: %v", allServers, err)
	}

	healthyServers, err := alg.HealthyServers()
	if err != nil || len(healthyServers) != len(params.Servers) {
		t.Errorf("HealthyServers returned unexpected result: %v, error: %v", healthyServers, err)
	}

	returned := make(map[string]bool)

	for i := 0; i < len(params.Servers); i++ {
		nextServer, err := alg.NextServer()
		if err != nil || nextServer == nil {
			t.Errorf("NextServer returned unexpected result: %v, error: %v", nextServer, err)
			continue
		}

		url := nextServer.GetUrl()

		if returned[url] {
			t.Errorf("Duplicate server returned by NextServer: %s", url)
		} else {
			returned[url] = true
		}
	}

	for _, s := range params.Servers {
		url := s.GetUrl()
		if !returned[url] {
			t.Errorf("Expected server not returned: %s", url)
		}
	}
}
