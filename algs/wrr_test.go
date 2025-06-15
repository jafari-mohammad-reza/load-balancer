package algs

import (
	"os"
	"testing"
)

func TestWeightedRoundRobin(t *testing.T) {
	os.Setenv("RUN_TYPE", "test")
	params := AlgParams{
		Servers: []IBackendServer{
			NewBackendServer("localhost", 8080, 3),
			NewBackendServer("localhost", 8081, 2),
			NewBackendServer("localhost", 8082, 1),
		},
	}
	alg, err := NewWeightedRoundRobinAlgorithm(params)
	if err != nil {
		t.Fatalf("Failed to create WeightedRoundRobinAlgorithm: %v", err)
	}
	allServers, err := alg.AllServers()
	if err != nil || len(allServers) != len(params.Servers) {
		t.Errorf("AllServers returned unexpected result: %v, error: %v", allServers, err)
	}

	healthyServers, err := alg.HealthyServers()
	if err != nil || len(healthyServers) != len(params.Servers) {
		t.Errorf("HealthyServers returned unexpected result: %v, error: %v", healthyServers, err)
	}
	expectedOrder := []string{"http://localhost:8080", "http://localhost:8080", "http://localhost:8080", "http://localhost:8081", "http://localhost:8081", "http://localhost:8082"}
	got := make([]string, 0, len(expectedOrder))

	for i := 0; i < len(expectedOrder); i++ {
		server, err := alg.NextServer()
		if err != nil {
			t.Fatalf("NextServer failed: %v", err)
		}
		got = append(got, server.GetUrl())
	}

	for i, url := range expectedOrder {
		if got[i] != url {
			t.Errorf("Expected url %s at index %d, but got %s", url, i, got[i])
		}
	}
}
