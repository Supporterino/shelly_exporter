package ShellySmartPlugS

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
	"github.com/supporterino/shelly_exporter/devices"
)

func RegisterSmartPlugS(device *devices.DeviceConfig, updateInterval time.Duration) {
	slog.Info("Registering Prometheus metrics")

	// Register all metrics with Prometheus
	prometheus.MustRegister(
		apiCounter,
		bleEnabledGauge,
		bleRPCEnabledGauge,
		bleObserverEnabledGauge,
		cloudEnabledGauge,
		mqttEnabledGauge,
		mqttRPCNotificationsGauge,
		mqttStatusNotificationsGauge,
		switchAutoOnGauge,
		switchAutoOnDelayGauge,
		switchPowerLimitGauge,
		switchVoltageLimitGauge,
		switchCurrentLimitGauge,
		wifiAPEnabledGauge,
		wifiSTAEnabledGauge,
		wifiRSSIThresholdGauge,
		cloudConnectedGauge,
		mqttConnectedGauge,
		switchOutputGauge,
		switchAPowerGauge,
		switchVoltageGauge,
		switchCurrentGauge,
		switchEnergyTotalGauge,
		switchTemperatureCGauge,
		sysUptimeGauge,
		sysRAMFreeGauge,
		sysFSFreeGauge,
		wifiRSSIGauge,
	)

	apiClient := client.NewAPIClient(device.Host, 10*time.Second)

	// Start fetching metrics periodically
	go func() {
		for {
			err := fetchAndUpdateMetrics(apiClient)
			if err != nil {
				slog.Error("Error fetching metrics", slog.Any("error", err))
			}
			time.Sleep(updateInterval * time.Second) // Adjust interval as needed
		}
	}()
}

// fetchAndUpdateMetrics fetches data from the API and updates Prometheus metrics.
func fetchAndUpdateMetrics(apiClient *client.APIClient) error {
	slog.Info("Fetching and updating metrics")

	err := fetchAndUpdateConfigMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update config metrics: %w", err)
	}

	err = fetchAndUpdateStatusMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update status metrics: %w", err)
	}

	slog.Info("Successfully updated metrics")
	return nil
}

// boolToFloat64 converts a boolean value to float64 (1 for true, 0 for false).
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
