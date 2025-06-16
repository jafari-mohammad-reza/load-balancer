package balancer

import (
	"fmt"
	"io"
	"load-balancer/algs"
	"load-balancer/conf"
	"load-balancer/log"
	"net/http"
	"path"
	"testing"
	"time"
)

func startMockServer(port int, body string) *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
	s := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: handler}
	go s.ListenAndServe()

	return s
}

func TestBalancer_ReverseProxy(t *testing.T) {

	s1 := startMockServer(9001, "server-1")
	s2 := startMockServer(9002, "server-2")
	time.Sleep(100 * time.Millisecond) // Give servers time to start
	defer s1.Close()
	defer s2.Close()
	tmpLogPath := t.TempDir()
	tmpPath := path.Join(tmpLogPath, "test-balancer.log")
	conf := &conf.Conf{
		Port:           9090,
		Algorithm:      "RoundRobin",
		BackendServers: []conf.BackendServer{conf.BackendServer{Host: "localhost", Port: 9001, Weight: 1}, conf.BackendServer{Host: "localhost", Port: 9002, Weight: 1}},
		Log: conf.LogConf{
			Logger:  conf.JSON,
			LogPath: tmpPath,
		},
	}
	alg, err := algs.NewAlgorithm(conf)
	if err != nil {
		t.Fatalf("Failed to initialize algorithm: %v", err)
	}
	logger, err := log.NewLogger(conf)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	b := NewBalancer(conf, alg, logger)

	go func() {
		err := b.ReverseProxy()
		if err != nil {
			t.Errorf("ReverseProxy failed: %v", err)
		}
	}()
	time.Sleep(200 * time.Millisecond)

	expected := []string{"server-1", "server-2", "server-1", "server-2"}
	for i, exp := range expected {
		resp, err := http.Get("http://localhost:9090")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if string(body) != exp {
			t.Errorf("Expected %q, got %q", exp, string(body))
		}
	}
}
