package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/supporterino/shelly_exporter/client"
	CoverGetStatus "github.com/supporterino/shelly_exporter/rpc/Cover.GetStatus"
	ShellyGetConfig "github.com/supporterino/shelly_exporter/rpc/Shelly.GetConfig"
	ShellyGetDeviceInfo "github.com/supporterino/shelly_exporter/rpc/Shelly.GetDeviceInfo"
	ShellyGetStatus "github.com/supporterino/shelly_exporter/rpc/Shelly.GetStatus"
	SwitchGetConfig "github.com/supporterino/shelly_exporter/rpc/Switch.GetConfig"
	SwitchGetStatus "github.com/supporterino/shelly_exporter/rpc/Switch.GetStatus"
	WiFiGetStatus "github.com/supporterino/shelly_exporter/rpc/WiFi.GetStatus"
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

	slog.Info("Registering new device", slog.String("host", device.Host))

	apiClient := client.NewAPIClient(device.Host, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize device type
	ShellyGetDeviceInfo.UpdateShellyGetDeviceInfoMetrics(apiClient)
	device.Type = ShellyGetDeviceInfo.GetDeviceType()
	device.Mac = ShellyGetDeviceInfo.GetDeviceMac()
	device.Profile = ShellyGetDeviceInfo.GetDeviceProfile()

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
				err := fetchAndUpdateMetrics(apiClient, device)
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
func fetchAndUpdateMetrics(apiClient *client.APIClient, device *DeviceConfig) error {
	slog.Info("Fetching and updating metrics")

	err := ShellyGetDeviceInfo.UpdateShellyGetDeviceInfoMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update device information metrics: %w", err)
	}

	err = ShellyGetStatus.UpdateShellyStatusMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update device information metrics: %w", err)
	}

	err = ShellyGetConfig.UpdateShellyGetConfigMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update config metrics: %w", err)
	}

	slog.Debug("Device type:", slog.String("type", device.Type))

	switch device_type := device.Type; device_type {
	case "Plus2PM":
		slog.Debug("Device profile:", slog.String("profile", device.Profile))
		switch profile := device.Profile; profile {
		case "cover":
			err := CoverGetStatus.UpdateCoverGetStatusMetrics(apiClient, 0, device.Mac)
			if err != nil {
				return fmt.Errorf("failed to update cover metrics: %w", err)
			}
			err = WiFiGetStatus.UpdateWiFiGetStatusMetrics(apiClient, device.Mac)
			if err != nil {
				return fmt.Errorf("failed to update wifi metrics: %w", err)
			}
		}
	case "PlusPlugS":
		err := SwitchGetStatus.UpdateSwitchGetStatusMetrics(apiClient, 0, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update switch status metrics: %w", err)
		}
		err = SwitchGetConfig.UpdateSwitchGetConfigMetrics(apiClient, 0, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update switch conig metrics: %w", err)
		}
		err = WiFiGetStatus.UpdateWiFiGetStatusMetrics(apiClient, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update wifi metrics: %w", err)
		}
	case "Mini1G3":
		err := SwitchGetStatus.UpdateSwitchGetStatusMetrics(apiClient, 0, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update switch status metrics: %w", err)
		}
		err = SwitchGetConfig.UpdateSwitchGetConfigMetrics(apiClient, 0, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update switch conig metrics: %w", err)
		}
		err = WiFiGetStatus.UpdateWiFiGetStatusMetrics(apiClient, device.Mac)
		if err != nil {
			return fmt.Errorf("failed to update wifi metrics: %w", err)
		}
	}

	slog.Info("Successfully updated metrics")
	return nil
}
