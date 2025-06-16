package balancer

import (
	"fmt"
	"load-balancer/algs"
	"load-balancer/conf"
	"load-balancer/log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type IBalancer interface {
	Start() error
}
type Balancer struct {
	conf       *conf.Conf
	logger     log.ILogger
	hostRouter map[int]map[string][]*routeHandler
}
type routeHandler struct {
	Path string
	Alg  algs.IAlgorithm
}

func (b *Balancer) Start() error {
	for _, proxy := range b.conf.Proxies {
		if err := b.registerProxy(proxy); err != nil {
			return err
		}
	}
	for port, hostMap := range b.hostRouter {
		p := port
		hm := hostMap
		go func() {
			b.logger.Info(fmt.Sprintf("Listening on :%d...", p))
			err := http.ListenAndServe(fmt.Sprintf(":%d", p), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b.routeRequest(w, r, p, hm)
			}))
			if err != nil {
				b.logger.Error(fmt.Sprintf("Failed to start listener on port %d: %v", p, err))
			}
		}()
	}
	select {}
}
func (b *Balancer) registerProxy(proxy conf.ProxyConf) error {
	if _, exists := b.hostRouter[proxy.Port]; !exists {
		b.hostRouter[proxy.Port] = make(map[string][]*routeHandler)
	}

	for _, loc := range proxy.Locations {
		alg, err := algs.NewAlgorithm(&loc)
		if err != nil {
			return fmt.Errorf("algorithm error on path %s: %w", loc.Path, err)
		}
		b.hostRouter[proxy.Port][proxy.Host] = append(b.hostRouter[proxy.Port][proxy.Host], &routeHandler{
			Path: loc.Path,
			Alg:  alg,
		})
	}
	return nil
}
func (b *Balancer) routeRequest(w http.ResponseWriter, r *http.Request, port int, hostMap map[string][]*routeHandler) {
	host := normalizeHost(r.Host)
	handlers, exists := hostMap[host]
	if !exists {
		http.Error(w, "host not found", http.StatusBadGateway)
		b.logger.Warn(fmt.Sprintf("No routes registered for host %s on port %d", host, port))
		return
	}

	for _, handler := range handlers {
		if strings.HasPrefix(r.URL.Path, handler.Path) {
			server, err := handler.Alg.NextServer()
			if err != nil {
				http.Error(w, "backend unavailable", http.StatusBadGateway)
				b.logger.Error(fmt.Sprintf("No backend for %s%s: %v", host, handler.Path, err))
				return
			}
			target, err := url.Parse(server.GetUrl())
			if err != nil {
				http.Error(w, "invalid backend url", http.StatusInternalServerError)
				b.logger.Error(fmt.Sprintf("Invalid backend URL %s: %v", server.GetUrl(), err))
				return
			}
			b.logger.Info(fmt.Sprintf("[%s] %s %s -> %s", host, r.Method, r.URL.Path, server.GetUrl()))
			httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "no matching route", http.StatusNotFound)
	b.logger.Warn(fmt.Sprintf("No matching path for %s%s", host, r.URL.Path))
}

func normalizeHost(raw string) string {
	if strings.Contains(raw, ":") {
		host, _, err := net.SplitHostPort(raw)
		if err == nil {
			return host
		}
	}
	return raw
}
func NewBalancer(conf *conf.Conf, logger log.ILogger) IBalancer {
	return &Balancer{
		conf:       conf,
		logger:     logger,
		hostRouter: make(map[int]map[string][]*routeHandler),
	}
}
