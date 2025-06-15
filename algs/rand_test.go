package algs

import (
	"os"
	"testing"
)

func TestRandomAlgorithm(t *testing.T) {
	os.Setenv("RUN_TYPE", "test")
	params := AlgParams{
		Servers: []IBackendServer{
			NewBackendServer("localhost", 8080, 1),
			NewBackendServer("localhost", 8081, 1),
			NewBackendServer("localhost", 8082, 1),
			NewBackendServer("localhost", 8083, 1),
		},
	}

	alg, err := NewRandomAlgorithm(params)
	if err != nil {
		t.Fatalf("Failed to create RandomAlgorithm: %v", err)
	}

	allServers, err := alg.AllServers()
	if err != nil || len(allServers) != len(params.Servers) {
		t.Errorf("AllServers returned unexpected result: %v, error: %v", allServers, err)
	}

	healthyServers, err := alg.HealthyServers()
	if err != nil || len(healthyServers) != len(params.Servers) {
		t.Errorf("HealthyServers returned unexpected result: %v, error: %v", healthyServers, err)
	}

	nextServer, err := alg.NextServer()
	if err != nil || nextServer == nil {
		t.Errorf("NextServer returned unexpected result: %v, error: %v", nextServer, err)
	}
}
