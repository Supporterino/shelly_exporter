package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register() {
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

	apiClient := client.NewAPIClient("http://10.1.255.111", 10*time.Second)

	// Start fetching metrics periodically
	go func() {
		for {
			err := fetchAndUpdateMetrics(apiClient)
			if err != nil {
				log.Printf("Error fetching metrics: %v", err)
			}
			time.Sleep(30 * time.Second) // Adjust interval as needed
		}
	}()
}

// fetchAndUpdateMetrics fetches data from the API and updates Prometheus metrics.
func fetchAndUpdateMetrics(apiClient *client.APIClient) error {
	err := fetchAndUpdateConfigMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update config metrics: %w", err)
	}

	err = fetchAndUpdateStatusMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update status metrics: %w", err)
	}

	return nil
}

// boolToFloat64 converts a boolean value to float64 (1 for true, 0 for false).
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
