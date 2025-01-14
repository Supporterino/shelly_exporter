package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporterino/shelly_exporter/metrics"
)

func main() {
	// Register custom metrics
	metrics.Register()

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting Prometheus exporter on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
