package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/supporterino/shelly_exporter/internal/config"
	"github.com/supporterino/shelly_exporter/internal/exporter"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Exporter failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	logger := exporter.NewLogger(cfg.Debug, os.Stdout)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := exporter.New(cfg, logger)
	if err != nil {
		return err
	}

	return app.Run(ctx)
}
