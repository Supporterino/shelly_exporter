package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/metrics"
)

func main() {
	// Configure slog based on the debug flag
	// var logger *slog.Logger
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	cfgPath, err := config.ParseFlags()
	if err != nil {
		logger.Error("Error parsing config path:", slog.Any("error", err))
	}
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		logger.Error("Error loading config:", slog.Any("error", err))
	}

	// Register custom metrics
	metrics.Register(cfg)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	logger.Info("Starting Prometheus exporter", slog.String("address", cfg.ListenAddress))
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Error("Error starting HTTP server", slog.Any("error", err))
	}
}
