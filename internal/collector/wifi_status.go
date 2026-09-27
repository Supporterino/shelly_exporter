package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// WiFiStatusCollector collects Wi-Fi metrics from WiFi.GetStatus.
type WiFiStatusCollector struct {
	Status *prometheus.GaugeVec
	SSID   *prometheus.GaugeVec
	RSSI   *prometheus.GaugeVec
}

// NewWiFiStatusCollector constructs and registers a WiFiStatusCollector.
func NewWiFiStatusCollector(reg prometheus.Registerer) (*WiFiStatusCollector, error) {
	c := &WiFiStatusCollector{
		Status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "status",
			Help:      "The status of the WiFi connection",
		}, []string{"device_mac", "status", "ip"}),
		SSID: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "ssid",
			Help:      "The SSID of the WiFi network",
		}, []string{"device_mac", "ssid"}),
		RSSI: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "rssi",
			Help:      "The Received Signal Strength Indicator (RSSI) of the WiFi connection in dBm",
		}, []string{"device_mac"}),
	}

	if err := register(reg, c.Status, c.SSID, c.RSSI); err != nil {
		return nil, fmt.Errorf("failed to register wifi status collector: %w", err)
	}

	return c, nil
}

// Delete removes every metric series for a device.
func (c *WiFiStatusCollector) Delete(deviceMAC string) {
	deleteDeviceSeries(deviceMAC, c.Status, c.SSID, c.RSSI)
}

// Update fetches WiFi.GetStatus and updates the collector metrics.
func (c *WiFiStatusCollector) Update(ctx context.Context, fetcher client.Fetcher, deviceMAC string) error {
	var status client.WiFiGetStatusResponse
	if err := fetcher.FetchData(ctx, "/rpc/WiFi.GetStatus", &status); err != nil {
		return fmt.Errorf("failed to fetch WiFi.GetStatus: %w", err)
	}

	c.UpdateMetrics(status, deviceMAC)
	return nil
}

// UpdateMetrics populates the metrics from a WiFi.GetStatus response.
func (c *WiFiStatusCollector) UpdateMetrics(status client.WiFiGetStatusResponse, deviceMAC string) {
	c.Status.WithLabelValues(deviceMAC, status.Status, status.StaIP).Set(1)
	c.SSID.WithLabelValues(deviceMAC, status.Ssid).Set(1)
	c.RSSI.WithLabelValues(deviceMAC).Set(float64(status.Rssi))
}
