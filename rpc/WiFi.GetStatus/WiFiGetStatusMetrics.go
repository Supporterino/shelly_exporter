package WiFiGetStatus

import (
	"fmt"

	"github.com/LukeEvansTech/shelly-prometheus-exporter/client"
	"github.com/LukeEvansTech/shelly-prometheus-exporter/labelset"
	"github.com/prometheus/client_golang/prometheus"
)

type WiFiGetStatusMetrics struct {
	Status *labelset.Gauge
	Ssid   *labelset.Gauge
	Rssi   *prometheus.GaugeVec
}

var metrics *WiFiGetStatusMetrics

func RegisterWiFiGetStatusMetrics() {
	metrics = &WiFiGetStatusMetrics{
		Status: labelset.New(prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "status",
			Help:      "The status of the WiFi connection",
		}, []string{"device_mac", "status", "ip"})),
		Ssid: labelset.New(prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "ssid",
			Help:      "The SSID of the WiFi network",
		}, []string{"device_mac", "ssid"})),
		Rssi: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "wifi",
			Name:      "rssi",
			Help:      "The Received Signal Strength Indicator (RSSI) of the WiFi connection",
		}, []string{"device_mac"}),
	}

	prometheus.MustRegister(
		metrics.Status,
		metrics.Ssid,
		metrics.Rssi,
	)
}

func UpdateWiFiGetStatusMetrics(apiClient *client.APIClient, deviceMac string) error {
	var config client.WiFiGetStatusResponse
	err := apiClient.FetchData("/rpc/WiFi.GetStatus", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config, deviceMac)

	return nil
}

func (m *WiFiGetStatusMetrics) UpdateMetrics(status client.WiFiGetStatusResponse, deviceMac string) {
	m.Status.Set(deviceMac, 1, deviceMac, status.Status, status.StaIP)
	m.Ssid.Set(deviceMac, 1, deviceMac, status.Ssid)
	// The value alone: an rssi LABEL made a new series on every change in
	// signal strength, several hundred per device over a few days.
	m.Rssi.WithLabelValues(deviceMac).Set(float64(status.Rssi))
}
