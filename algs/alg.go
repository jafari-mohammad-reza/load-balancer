package algs

import (
	"errors"
	"sync"
	"time"
)

type ServerStatus string

const (
	Healthy   ServerStatus = "Healthy"
	UnHealthy ServerStatus = "UnHealthy"
)

type IBackendServer interface {
	Ping() error
	SetStatus(status ServerStatus) error
	GetStatus() ServerStatus
}
type BackendServer struct {
	Host          string
	Port          int
	PingUrl       string
	ReqCount      int
	TotalReqCount int // the request count the server can handles
	Weight        int
	Status        ServerStatus
	LastChecked   time.Time
	mu            sync.RWMutex
}

func (s *BackendServer) Ping() error {
	return nil
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
func NewBackendServer(host string, port int, pingUrl string, totalReqCount int, weight int) *BackendServer {
	return &BackendServer{
		Host:          host,
		Port:          port,
		PingUrl:       pingUrl,
		ReqCount:      0,
		TotalReqCount: totalReqCount,
		Weight:        weight,
		Status:        Healthy,
		LastChecked:   time.Now(),
		mu:            sync.RWMutex{},
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
