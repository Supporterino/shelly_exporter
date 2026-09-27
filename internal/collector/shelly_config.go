package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// ShellyConfigCollector collects device configuration metrics from Shelly.GetConfig.
type ShellyConfigCollector struct {
	BLEEnabled           *prometheus.GaugeVec
	CloudEnabled         *prometheus.GaugeVec
	CloudServer          *prometheus.GaugeVec
	EthEnabled           *prometheus.GaugeVec
	EthIPv4Mode          *prometheus.GaugeVec
	WiFiAPEnabled        *prometheus.GaugeVec
	WiFiSTAEnabled       *prometheus.GaugeVec
	WiFiRoamingThreshold *prometheus.GaugeVec
}

// NewShellyConfigCollector constructs and registers a ShellyConfigCollector.
func NewShellyConfigCollector(reg prometheus.Registerer) (*ShellyConfigCollector, error) {
	c := &ShellyConfigCollector{
		BLEEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "ble",
			Help:      "Indicates if BLE is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		CloudEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "cloud",
			Help:      "Indicates if Cloud is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		CloudServer: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "cloud_server",
			Help:      "Cloud server configuration (labels include server address)",
		}, []string{"device_mac", "server"}),
		EthEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "eth",
			Help:      "Indicates if Ethernet is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		EthIPv4Mode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "eth_ipv4_mode",
			Help:      "Ethernet IPv4 mode (labels include mode)",
		}, []string{"device_mac", "mode"}),
		WiFiAPEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_ap",
			Help:      "Indicates if Wi-Fi AP is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WiFiSTAEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_sta",
			Help:      "Indicates if Wi-Fi STA is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WiFiRoamingThreshold: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_roaming_rssi_threshold",
			Help:      "RSSI threshold for Wi-Fi roaming",
		}, []string{"device_mac"}),
	}

	if err := register(reg,
		c.BLEEnabled,
		c.CloudEnabled,
		c.CloudServer,
		c.EthEnabled,
		c.EthIPv4Mode,
		c.WiFiAPEnabled,
		c.WiFiSTAEnabled,
		c.WiFiRoamingThreshold,
	); err != nil {
		return nil, fmt.Errorf("failed to register shelly config collector: %w", err)
	}

	return c, nil
}

// Delete removes every metric series for a device.
func (c *ShellyConfigCollector) Delete(deviceMAC string) {
	deleteDeviceSeries(deviceMAC,
		c.BLEEnabled, c.CloudEnabled, c.CloudServer, c.EthEnabled,
		c.EthIPv4Mode, c.WiFiAPEnabled, c.WiFiSTAEnabled, c.WiFiRoamingThreshold,
	)
}

// Update fetches Shelly.GetConfig, updates the collector metrics, and returns
// the decoded configuration so callers can discover the device's components.
func (c *ShellyConfigCollector) Update(ctx context.Context, fetcher client.Fetcher) (client.ShellyGetConfigResponse, error) {
	var cfg client.ShellyGetConfigResponse
	if err := fetcher.FetchData(ctx, "/rpc/Shelly.GetConfig", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to fetch Shelly.GetConfig: %w", err)
	}

	c.UpdateMetrics(cfg)
	return cfg, nil
}

// UpdateMetrics populates the metrics from a Shelly.GetConfig response.
func (c *ShellyConfigCollector) UpdateMetrics(cfg client.ShellyGetConfigResponse) {
	mac := cfg.Sys.Device.MAC

	c.BLEEnabled.WithLabelValues(mac).Set(boolToFloat64(cfg.BLE.Enable))
	c.CloudEnabled.WithLabelValues(mac).Set(boolToFloat64(cfg.Cloud.Enable))
	c.CloudServer.WithLabelValues(mac, cfg.Cloud.Server).Set(1)

	c.EthEnabled.WithLabelValues(mac).Set(boolToFloat64(cfg.Eth.Enable))
	if cfg.Eth.IPv4Mode != "" {
		c.EthIPv4Mode.WithLabelValues(mac, cfg.Eth.IPv4Mode).Set(1)
	}

	c.WiFiAPEnabled.WithLabelValues(mac).Set(boolToFloat64(cfg.Wifi.AP.Enable))
	c.WiFiSTAEnabled.WithLabelValues(mac).Set(boolToFloat64(cfg.Wifi.STA.Enable))
	c.WiFiRoamingThreshold.WithLabelValues(mac).Set(float64(cfg.Wifi.Roam.RSSIThreshold))
}
