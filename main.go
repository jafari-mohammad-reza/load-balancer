package main

import (
	"fmt"
	"load-balancer/algs"
	"load-balancer/balancer"
	"load-balancer/conf"
	"load-balancer/log"
	"os"
)

func main() {
	conf, err := conf.ReadConf()
	if err != nil {
		fmt.Printf("error reading conf %v", err)
		os.Exit(1)
	}
	logger, err := log.NewLogger(conf)
	if err != nil {
		fmt.Printf("error creating logger %v", err)
		os.Exit(1)
	}
	alg, err := algs.NewAlgorithm(conf)
	if err != nil {
		logger.Error(fmt.Errorf("error creating algorithm %v", err))
		os.Exit(1)
	}
	balancer := balancer.NewBalancer(conf, alg, logger)
	if err := balancer.ReverseProxy(); err != nil {
		logger.Error(fmt.Errorf("error starting reverse proxy: %v", err))
		os.Exit(1)
	}
}
