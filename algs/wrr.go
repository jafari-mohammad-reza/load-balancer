package algs

type WeightedRoundRobinAlgorithm struct {
	Servers        []IBackendServer
	healthyServers []IBackendServer
	CurrentIndex   int
}

func (r *WeightedRoundRobinAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *WeightedRoundRobinAlgorithm) HealthyServers() ([]IBackendServer, error) {
	return r.healthyServers, nil
}
func (r *WeightedRoundRobinAlgorithm) NextServer() (IBackendServer, error) {
	return nil, nil
}

func NewWeightedRoundRobinAlgorithm(params AlgParams) (*WeightedRoundRobinAlgorithm, error) {
	return &WeightedRoundRobinAlgorithm{
		Servers:      params.Servers,
		CurrentIndex: -1,
	}, nil
}
