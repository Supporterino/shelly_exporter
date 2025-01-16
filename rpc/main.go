package rpc

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/supporterino/shelly_exporter/client"
	ShellyGetConfig "github.com/supporterino/shelly_exporter/rpc/Shelly.GetConfig"
	ShellyGetDeviceInfo "github.com/supporterino/shelly_exporter/rpc/Shelly.GetDeviceInfo"
	ShellyGetStatus "github.com/supporterino/shelly_exporter/rpc/Shelly.GetStatus"
)

func RegisterDevice(device *DeviceConfig, updateInterval time.Duration) {
	slog.Info("Registering Prometheus metrics")

	ShellyGetConfig.RegisterShelly_GetConfigMetrics()
	ShellyGetStatus.RegisterShelly_GetStatusMetrics()
	ShellyGetDeviceInfo.RegisterShelly_GetDeviceInfoMetrics()

	apiClient := client.NewAPIClient(device.Host, 10*time.Second)

	// Start fetching metrics periodically
	go func() {
		for {
			time.Sleep(updateInterval * time.Second) // Adjust interval as needed
			err := fetchAndUpdateMetrics(apiClient)
			if err != nil {
				slog.Error("Error fetching metrics", slog.Any("error", err))
			}
		}
	}()
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
