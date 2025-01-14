package main

import (
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/metrics"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v", err)
	}

	// Register custom metrics
	metrics.Register(&cfg)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting Prometheus exporter on :8080")
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
