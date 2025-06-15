package balancer

import (
	"fmt"
	"load-balancer/algs"
	"load-balancer/conf"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type IBalancer interface {
	ReverseProxy() error
}
type Balancer struct {
	conf *conf.Conf
	alg  algs.IAlgorithm
}

func (b *Balancer) ReverseProxy() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server, err := b.alg.NextServer()
		if err != nil {
			fmt.Printf("error while selecting next server %v", err)
			return
		}
		serverUrl, err := url.Parse(server.GetUrl())
		if err != nil {
			fmt.Printf("error while parsing server url %s -  %v", server.GetUrl(), err)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Listening on :%d...", b.conf.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", b.conf.Port), nil); err != nil {
		return fmt.Errorf("server failed: %v", err)
	}
	return nil
}
func NewBalancer(conf *conf.Conf, alg algs.IAlgorithm) IBalancer {
	return &Balancer{
		conf: conf,
		alg:  alg,
	}
}
