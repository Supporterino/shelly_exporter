package main

import (
	"log/slog"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/metrics"
)

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Error parsing config", slog.Any("error", err))
	}

	// Configure slog based on the debug flag
	var logger *slog.Logger
	if cfg.Debug {
		logger = slog.New(slog.NewTextHandler(slog.StderrHandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.NewTextHandler(slog.StderrHandlerOptions{Level: slog.LevelInfo}))
	}
	slog.SetDefault(logger)

	// Register custom metrics
	metrics.Register(&cfg)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	logger.Info("Starting Prometheus exporter", slog.String("address", cfg.ListenAddress))
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Error("Error starting HTTP server", slog.Any("error", err))
	}
}
