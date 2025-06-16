package balancer

import (
	"fmt"
	"io"

	"load-balancer/conf"
	"load-balancer/log"
	"net/http"
	"path"
	"testing"
)

func startMockServer(port int, body string) *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: handler}
	go srv.ListenAndServe()
	return srv
}

func TestBalancer_Start(t *testing.T) {
	s8001 := startMockServer(8001, "from-8001")
	s8002 := startMockServer(8002, "from-8002")
	s9001 := startMockServer(9001, "from-9001")
	defer s8001.Close()
	defer s8002.Close()
	defer s9001.Close()

	cfg := &conf.Conf{
		Proxies: []conf.ProxyConf{
			{
				Port: 8080,
				Host: "example.com",
				Locations: []conf.LocationConf{
					{
						Path:      "/api",
						Algorithm: "RoundRobin",
						BackendServers: []conf.BackendServer{
							{Host: "localhost", Port: 8001},
							{Host: "localhost", Port: 8002},
						},
					},
				},
			},
			{
				Port: 9090,
				Host: "another.com",
				Locations: []conf.LocationConf{
					{
						Path:      "/",
						Algorithm: "Random",
						BackendServers: []conf.BackendServer{
							{Host: "localhost", Port: 9001},
						},
					},
				},
			},
		},
	}

	logPath := path.Join(t.TempDir(), "test.log")
	cfg.Log = conf.LogConf{
		Logger:  conf.JSON,
		LogPath: logPath,
	}
	logger, err := log.NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	b := NewBalancer(cfg, logger)
	go func() {
		if err := b.Start(); err != nil {
			t.Errorf("balancer.Start() error: %v", err)
		}
	}()

	results := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", "http://localhost:8080/api", nil)
		req.Host = "example.com"
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request to roundrobin failed: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		results = append(results, string(body))
	}
	if results[0] == results[1] && results[1] == results[2] {
		t.Errorf("RoundRobin failed, all responses are the same: %v", results)
	}

	req, _ := http.NewRequest("GET", "http://localhost:9090/", nil)
	req.Host = "another.com"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request to random failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "from-9001" {
		t.Errorf("Expected 'from-9001', got %s", body)
	}
}
