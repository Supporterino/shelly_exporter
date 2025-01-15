package main

import (
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

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	logger.Info("Starting Prometheus exporter", slog.String("address", cfg.ListenAddress))
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Error("Error starting HTTP server", slog.Any("error", err))
	}
}
