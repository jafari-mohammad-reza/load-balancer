package algs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ServerStatus string

const (
	Healthy   ServerStatus = "Healthy"
	UnHealthy ServerStatus = "UnHealthy"
)

type IBackendServer interface {
	SetStatus(status ServerStatus) error
	GetStatus() ServerStatus
	GetID() uuid.UUID
	GetUrl() string
	IncrementReqCount() int
	SetWeight(weight int) error
	GetWeight() int
}
type BackendServer struct {
	ID          uuid.UUID
	Host        string
	Port        int
	ReqCount    int
	Weight      int
	Status      ServerStatus
	LastChecked time.Time
	mu          sync.RWMutex
}

func (s *BackendServer) SetStatus(status ServerStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
	return nil
}
func (s *BackendServer) GetStatus() ServerStatus {
	return s.Status
}
func (s *BackendServer) GetID() uuid.UUID {
	return s.ID
}
func (s *BackendServer) GetUrl() string {
	return fmt.Sprintf("http://%s:%d", s.Host, s.Port)
}
func (s *BackendServer) IncrementReqCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReqCount++
	return s.ReqCount
}
func (s *BackendServer) SetWeight(weight int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if weight < 0 {
		return errors.New("weight cannot be negative")
	}
	s.Weight = weight
	return nil
}
func (s *BackendServer) GetWeight() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Weight
}

func NewBackendServer(host string, port int, weight int) *BackendServer {
	return &BackendServer{
		ID:          uuid.New(),
		Host:        host,
		Port:        port,
		ReqCount:    0,
		Weight:      weight,
		Status:      Healthy,
		LastChecked: time.Now(),
		mu:          sync.RWMutex{},
	}
}

type IAlgorithm interface {
	AllServers() ([]IBackendServer, error)
	HealthyServers() ([]IBackendServer, error)
	NextServer() (IBackendServer, error)
}
type Alg string

const (
	Random             Alg = "random"
	RoundRobin         Alg = "RoundRobin"
	WeightedRoundRobin Alg = "WeightedRoundRobin"
)

type AlgParams struct {
	Servers []IBackendServer
}

func NewAlgorithm(alg Alg, params AlgParams) (IAlgorithm, error) {
	switch alg {
	case Random:
		return NewRandomAlgorithm(params)
	case RoundRobin:
		return NewRoundRobinAlgorithm(params)
	case WeightedRoundRobin:
		return NewWeightedRoundRobinAlgorithm(params)
	default:
		return nil, errors.New("unsupported algorithm")
	}
}

func Ping(server IBackendServer) error {
	if os.Getenv("RUN_TYPE") == "test" {
		return nil
	}
	url := fmt.Sprintf("%s/ping", server.GetUrl())
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Request failed:", err)
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return err
	}
	defer resp.Body.Close()
	if string(body) != "Pong" {
		fmt.Println("Unexpected response:", string(body))
		return err
	}
	return errors.New(":ping error")
}
