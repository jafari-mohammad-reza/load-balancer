package algs

type RandomAlgorithm struct {
	Servers        []IBackendServer
	healthyServers []IBackendServer
	CurrentIndex   int
}

func (r *RandomAlgorithm) AllServers() ([]IBackendServer, error) {
	return r.Servers, nil
}
func (r *RandomAlgorithm) HealthyServers() ([]IBackendServer, error) {
	return r.healthyServers, nil
}
func (r *RandomAlgorithm) NextServer() (IBackendServer, error) {
	return nil, nil
}

func NewRandomAlgorithm(params AlgParams) (*RandomAlgorithm, error) {
	return &RandomAlgorithm{
		Servers:      params.Servers,
		CurrentIndex: -1,
	}, nil
}
