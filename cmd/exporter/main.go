package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/metrics"
)

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal("Error parsing config path:", slog.Any("error", err))
	}
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal("Error loading config:", slog.Any("error", err))
	}

	// Configure slog based on the debug flag
	var logger *slog.Logger
	if cfg.Debug {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	slog.SetDefault(logger)

	// Register custom metrics
	metrics.Register(cfg)

	// Expose endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Shelly Exporter</title></head>
             <body>
             <h1>Haproxy Exporter</h1>
             <p><a href=/metrics>Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", healthHandler)

	logger.Info("Starting Prometheus exporter", slog.String("address", cfg.ListenAddress))
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Error("Error starting HTTP server", slog.Any("error", err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Check the health of the server and return a status code accordingly
	if serverIsHealthy() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Server is healthy")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Server is not healthy")
	}
}

func serverIsHealthy() bool {
	// Check the health of the server and return true or false accordingly
	// For example, check if the server can connect to the database
	return true
}
