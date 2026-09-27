package exporter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/supporterino/shelly_exporter/internal/config"
)

const shutdownTimeout = 10 * time.Second

// App wires configuration, collectors, devices, and the HTTP server together.
type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	registry   *prometheus.Registry
	collectors *Collectors
	manager    *DeviceManager
	server     *http.Server
}

// New builds an App from configuration and a logger.
func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.Default()
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	cols, err := NewCollectors(registry)
	if err != nil {
		return nil, err
	}

	manager := NewDeviceManager(cfg.DeviceUpdateInterval.Duration(), cols, DefaultFetcherFactory, logger)

	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           NewHandler(registry),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		registry:   registry,
		collectors: cols,
		manager:    manager,
		server:     server,
	}, nil
}

// Run starts device polling and the HTTP server, blocking until ctx is
// cancelled or the server fails.
func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, device := range a.cfg.Devices {
		a.manager.Register(ctx, Device{
			Host:     device.Host,
			Username: device.Username,
			Password: device.Password,
		})
	}

	serverErr := make(chan error, 1)
	go func() {
		a.logger.Info("Starting Prometheus exporter", slog.String("address", a.cfg.ListenAddress))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-serverErr:
	}

	a.logger.Info("Shutting down exporter")

	// Stop the polling loops even when the server fails before ctx is
	// cancelled, so that manager.Wait can join them.
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	shutdownErr := a.server.Shutdown(shutdownCtx)
	a.manager.Wait()

	if runErr != nil {
		return runErr
	}
	if shutdownErr != nil {
		return fmt.Errorf("server shutdown failed: %w", shutdownErr)
	}
	return nil
}
