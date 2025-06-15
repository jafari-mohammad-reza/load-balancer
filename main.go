package main

import (
	"fmt"
	"load-balancer/algs"
	"load-balancer/balancer"
	"load-balancer/conf"
	"os"
)

func main() {
	conf, err := conf.ReadConf()
	if err != nil {
		fmt.Printf("error reading conf %v", err)
		os.Exit(1)
	}
	alg, err := algs.NewAlgorithm(conf)
	if err != nil {
		fmt.Printf("error creating algorithm %v", err)
		os.Exit(1)
	}
	balancer := balancer.NewBalancer(conf, alg)
	if err := balancer.ReverseProxy(); err != nil {
		fmt.Printf("error starting reverse proxy: %v\n", err)
		os.Exit(1)
	}
}
