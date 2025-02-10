package ShellyGetStatus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type ShellyGetStatusMetrics struct {
	Uptime   *prometheus.GaugeVec
	RAM      *prometheus.GaugeVec
	FS       *prometheus.GaugeVec
	WIFIRSSI *prometheus.GaugeVec
}

var metrics *ShellyGetStatusMetrics

func RegisterShellyGetStatusMetrics() {
	metrics = &ShellyGetStatusMetrics{
		Uptime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "system",
			Name:      "uptime",
			Help:      "System uptime in seconds",
		}, []string{"device_mac"}),
		RAM: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "system",
			Name:      "ram",
			Help:      "RAM sizes free and used in bytes",
		}, []string{"device_mac", "kind"}),
		FS: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "system",
			Name:      "fs",
			Help:      "FS sizes free and used in bytes",
		}, []string{"device_mac", "kind"}),
		WIFIRSSI: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "system",
			Name:      "wifi_rssi",
			Help:      "Wi-Fi RSSI signal strength in dBm",
		}, []string{"device_mac", "ssid", "sta_ip"}),
	}

	prometheus.MustRegister(
		metrics.Uptime,
		metrics.FS,
		metrics.RAM,
		metrics.WIFIRSSI,
	)
}

func UpdateShellyStatusMetrics(apiClient *client.APIClient) error {
	var config client.ShellyGetStatusResponse
	err := apiClient.FetchData("/rpc/Shelly.GetStatus", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config)

	return nil
}

func (m *ShellyGetStatusMetrics) UpdateMetrics(status client.ShellyGetStatusResponse) {
	deviceMAC := status.Sys.MAC

	m.Uptime.WithLabelValues(deviceMAC).Set(float64(status.Sys.Uptime))
	m.RAM.WithLabelValues(deviceMAC, "free").Set(float64(status.Sys.RAMFree))
	m.RAM.WithLabelValues(deviceMAC, "max").Set(float64(status.Sys.RAMSize))
	m.FS.WithLabelValues(deviceMAC, "free").Set(float64(status.Sys.FSFree))
	m.FS.WithLabelValues(deviceMAC, "max").Set(float64(status.Sys.FSSize))
	m.WIFIRSSI.WithLabelValues(deviceMAC, *status.Wifi.SSID, *status.Wifi.StaIP).Set(float64(status.Wifi.RSSI))
}
