package balancer

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"load-balancer/algs"
	"load-balancer/conf"
	"load-balancer/log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
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
			return fmt.Errorf("failed to register proxy for host %s: %w", proxy.Host, err)
		}
	}

	for port, hostMap := range b.hostRouter {
		proxyConf, ok := b.getProxyByPort(port)
		if !ok {
			b.logger.Error(fmt.Sprintf("No proxy configuration found for port %d", port))
			continue
		}

		go func(port int, hostMap map[string][]*routeHandler, proxy conf.ProxyConf) {
			addr := fmt.Sprintf(":%d", port)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b.routeRequest(w, r, port, hostMap)
			})

			server := &http.Server{
				Addr:    addr,
				Handler: handler,
			}

			if proxy.TLS {
				b.logger.Info(fmt.Sprintf("Listening with TLS on %s", addr))

				tlsConfig, err := b.buildTLSConfig(proxy)
				if err != nil {
					b.logger.Error(fmt.Sprintf("Failed to build TLS config for port %d: %v", port, err))
					return
				}
				server.TLSConfig = tlsConfig

				if err := server.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
					b.logger.Error(fmt.Sprintf("HTTPS server error on port %d: %v", port, err))
				}
			} else {
				b.logger.Info(fmt.Sprintf("Listening on %s", addr))

				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					b.logger.Error(fmt.Sprintf("HTTP server error on port %d: %v", port, err))
				}
			}
		}(port, hostMap, proxyConf)
	}

	select {}
}
func (b *Balancer) getProxyByPort(port int) (conf.ProxyConf, bool) {
	for _, proxy := range b.conf.Proxies {
		if proxy.Port == port {
			return proxy, true
		}
	}
	return conf.ProxyConf{}, false
}

func (b *Balancer) buildTLSConfig(proxy conf.ProxyConf) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(proxy.Certificate, proxy.CertificateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert/key: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCert, err := os.ReadFile(proxy.ClientCA)
	if err != nil {
		return nil, fmt.Errorf("failed to read client CA cert: %w", err)
	}
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
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
	cleanPath := path.Clean(r.URL.Path)
	for _, handler := range handlers {
		if strings.HasPrefix(cleanPath, handler.Path) {
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
			b.logger.Info(fmt.Sprintf("[%s] %s %s -> %s", host, r.Method, cleanPath, server.GetUrl()))
			httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "no matching route", http.StatusNotFound)
	b.logger.Warn(fmt.Sprintf("No matching path for %s%s", host, cleanPath))
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
