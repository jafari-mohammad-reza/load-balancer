package algs

type RoundRobinAlgorithm struct {
	Servers        []IBackendServer
	healthyServers []IBackendServer
	CurrentIndex   int
}

func (r *RoundRobinAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *RoundRobinAlgorithm) HealthyServers() ([]IBackendServer, error) {
	return r.healthyServers, nil
}
func (r *RoundRobinAlgorithm) NextServer() (IBackendServer, error) {
	return nil, nil
}

func NewRoundRobinAlgorithm(params AlgParams) (*RoundRobinAlgorithm, error) {
	return &RoundRobinAlgorithm{
		Servers:      params.Servers,
		CurrentIndex: -1,
	}, nil
}
