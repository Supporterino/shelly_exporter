package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/supporterino/shelly_exporter/client"
	ShellyGetConfig "github.com/supporterino/shelly_exporter/rpc/Shelly.GetConfig"
	ShellyGetDeviceInfo "github.com/supporterino/shelly_exporter/rpc/Shelly.GetDeviceInfo"
	ShellyGetStatus "github.com/supporterino/shelly_exporter/rpc/Shelly.GetStatus"
)

// DeviceManager manages registered devices.
type DeviceManager struct {
	mu      sync.Mutex
	devices map[string]context.CancelFunc
}

// NewDeviceManager creates a new DeviceManager.
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]context.CancelFunc),
	}
}

// RegisterDevice registers a device and starts its metrics update loop.
func (dm *DeviceManager) RegisterDevice(device *DeviceConfig, updateInterval time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.devices[device.Host]; exists {
		slog.Warn("Device already registered", slog.String("host", device.Host))
		return
	}

	slog.Info("Registering Prometheus metrics", slog.String("host", device.Host))

	apiClient := client.NewAPIClient(device.Host, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Save the cancel function to stop the goroutine later.
	dm.devices[device.Host] = cancel

	// Start fetching metrics periodically
	go func() {
		ticker := time.NewTicker(updateInterval * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Stopping metrics update loop", slog.String("host", device.Host))
				return
			case <-ticker.C:
				err := fetchAndUpdateMetrics(apiClient)
				if err != nil {
					slog.Error("Error fetching metrics", slog.Any("error", err), slog.String("host", device.Host))
				}
			}
		}
	}()
}

// DeregisterDevice stops the metrics update loop for a device.
func (dm *DeviceManager) DeregisterDevice(host string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	cancel, exists := dm.devices[host]
	if !exists {
		slog.Warn("Device not found", slog.String("host", host))
		return
	}

	// Call the cancel function to stop the loop.
	cancel()
	delete(dm.devices, host)
	slog.Info("Device deregistered", slog.String("host", host))
}

// DeregisterAll stops all metrics update loops and clears the device list.
func (dm *DeviceManager) DeregisterAll() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for host, cancel := range dm.devices {
		slog.Info("Deregistering device", slog.String("host", host))
		cancel()
	}
	dm.devices = make(map[string]context.CancelFunc)
	slog.Info("All devices deregistered")
}

// fetchAndUpdateMetrics fetches data from the API and updates Prometheus metrics.
func fetchAndUpdateMetrics(apiClient *client.APIClient) error {
	slog.Info("Fetching and updating metrics")

	err := ShellyGetDeviceInfo.UpdateShelly_GetDeviceInfoMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update information metrics: %w", err)
	}

	err = ShellyGetConfig.UpdateShelly_GetConfigMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update config metrics: %w", err)
	}

	err = ShellyGetStatus.UpdateShelly_StatusMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update status metrics: %w", err)
	}

	slog.Info("Successfully updated metrics")
	return nil
}
