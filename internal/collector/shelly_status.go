package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// ShellyStatusCollector collects system status metrics from Shelly.GetStatus.
type ShellyStatusCollector struct {
	Uptime   *prometheus.GaugeVec
	RAM      *prometheus.GaugeVec
	FS       *prometheus.GaugeVec
	WiFiRSSI *prometheus.GaugeVec
}

// NewShellyStatusCollector constructs and registers a ShellyStatusCollector.
func NewShellyStatusCollector(reg prometheus.Registerer) (*ShellyStatusCollector, error) {
	c := &ShellyStatusCollector{
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
		WiFiRSSI: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "system",
			Name:      "wifi_rssi",
			Help:      "Wi-Fi RSSI signal strength in dBm",
		}, []string{"device_mac", "ssid", "sta_ip"}),
	}

	if err := register(reg, c.Uptime, c.RAM, c.FS, c.WiFiRSSI); err != nil {
		return nil, fmt.Errorf("failed to register shelly status collector: %w", err)
	}

	return c, nil
}

// Update fetches Shelly.GetStatus and updates the collector metrics.
func (c *ShellyStatusCollector) Update(ctx context.Context, fetcher client.Fetcher) error {
	var status client.ShellyGetStatusResponse
	if err := fetcher.FetchData(ctx, "/rpc/Shelly.GetStatus", &status); err != nil {
		return fmt.Errorf("failed to fetch Shelly.GetStatus: %w", err)
	}

	c.UpdateMetrics(status)
	return nil
}

// UpdateMetrics populates the metrics from a Shelly.GetStatus response.
func (c *ShellyStatusCollector) UpdateMetrics(status client.ShellyGetStatusResponse) {
	deviceMAC := status.Sys.MAC

	c.Uptime.WithLabelValues(deviceMAC).Set(float64(status.Sys.Uptime))
	c.RAM.WithLabelValues(deviceMAC, "free").Set(float64(status.Sys.RAMFree))
	c.RAM.WithLabelValues(deviceMAC, "max").Set(float64(status.Sys.RAMSize))
	c.FS.WithLabelValues(deviceMAC, "free").Set(float64(status.Sys.FSFree))
	c.FS.WithLabelValues(deviceMAC, "max").Set(float64(status.Sys.FSSize))

	ssid := ""
	if status.Wifi.SSID != nil {
		ssid = *status.Wifi.SSID
	}
	staIP := ""
	if status.Wifi.StaIP != nil {
		staIP = *status.Wifi.StaIP
	}
	c.WiFiRSSI.WithLabelValues(deviceMAC, ssid, staIP).Set(float64(status.Wifi.RSSI))
}
