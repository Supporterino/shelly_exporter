package ShellySmartPlugS

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

var (
	// Device Info Metrics
	deviceInfoGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_device_info",
		Help: "Static device information exposed as labels (model, firmware version, app).",
	}, []string{"device_name", "device_id", "device_mac", "model", "fw_version", "app"})

	authEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_auth_enabled",
		Help: "Indicates if authentication is enabled on the device.",
	}, []string{"device_mac"})
)

func fetchAndUpdateDeviceInfo(apiClient *client.APIClient) error {
	var deviceInfo client.ShellyGetDeviceInfoResponse
	err := apiClient.FetchData("/rpc/Shelly.GetDeviceInfo", &deviceInfo)
	if err != nil {
		return fmt.Errorf("error fetching device info: %w", err)
	}

	updateDeviceInfoMetrics(deviceInfo)
	return nil
}

func updateDeviceInfoMetrics(deviceInfo client.ShellyGetDeviceInfoResponse) {
	// Update device info metric with labels
	deviceInfoGauge.With(prometheus.Labels{
		"device_name": deviceInfo.Name,
		"device_id":   deviceInfo.ID,
		"device_mac":  deviceInfo.MAC,
		"model":       deviceInfo.Model,
		"fw_version":  deviceInfo.Version,
		"app":         deviceInfo.App,
	}).Set(1) // Set a static value for the gauge

	// Update authentication status metric
	authEnabledGauge.With(prometheus.Labels{
		"device_mac": deviceInfo.MAC,
	}).Set(boolToFloat64(deviceInfo.AuthEn))
}
