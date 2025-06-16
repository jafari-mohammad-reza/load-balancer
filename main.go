package main

import (
	"fmt"
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
	balancer := balancer.NewBalancer(conf, logger)
	if err := balancer.Start(); err != nil {
		logger.Error(fmt.Errorf("error starting reverse proxy: %v", err))
		os.Exit(1)
	}
}
