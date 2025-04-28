package WiFiGetStatus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type WiFiGetStatusMetrics struct {
	Status *prometheus.GaugeVec
	Ssid   *prometheus.GaugeVec
	Rssi   *prometheus.GaugeVec
}

var metrics *WiFiGetStatusMetrics

func RegisterWiFiGetStatusMetrics() {
	metrics = &WiFiGetStatusMetrics{
		Status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "status",
			Help:      "The status of the WiFi connection",
		}, []string{"device_mac", "status", "ip"}),
		Ssid: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "ssid",
			Help:      "The SSID of the WiFi network",
		}, []string{"device_mac", "ssid"}),
		Rssi: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "rssi",
			Help:      "The Received Signal Strength Indicator (RSSI) of the WiFi connection",
		}, []string{"device_mac", "rssi"}),
	}

	prometheus.MustRegister(
		metrics.Status,
		metrics.Ssid,
		metrics.Rssi,
	)
}

func UpdateWiFiGetStatusMetrics(apiClient *client.APIClient, device_mac string) error {
	var config client.WiFiGetStatusResponse
	err := apiClient.FetchData("/rpc/WiFi.GetStatus", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config, device_mac)

	return nil
}

func (m *WiFiGetStatusMetrics) UpdateMetrics(status client.WiFiGetStatusResponse, device_mac string) {
	m.Status.WithLabelValues(device_mac, status.Status, status.StaIP).Set(1)
	m.Ssid.WithLabelValues(device_mac, status.Ssid).Set(1)
	m.Rssi.WithLabelValues(device_mac, fmt.Sprintf("%d", status.Rssi)).Set(float64(status.Rssi))
}
