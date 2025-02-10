package ShellyGetConfig

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type ConfigMetrics struct {
	BLEEnabled           *prometheus.GaugeVec
	CloudEnabled         *prometheus.GaugeVec
	CloudServer          *prometheus.GaugeVec
	EthEnabled           *prometheus.GaugeVec
	EthIPv4Mode          *prometheus.GaugeVec
	WifiAPEnabled        *prometheus.GaugeVec
	WifiSTAEnabled       *prometheus.GaugeVec
	WifiRoamingThreshold *prometheus.GaugeVec
}

var metrics *ConfigMetrics

// RegisterMetrics initializes and registers the Prometheus metrics
func RegisterShellyGetConfigMetrics() {
	metrics = &ConfigMetrics{
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
			Namespace: "selly",
			Subsystem: "device",
			Name:      "eth_ipv4_mode",
			Help:      "Ethernet IPv4 mode (labels include mode)",
		}, []string{"device_mac", "mode"}),
		WifiAPEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_ap",
			Help:      "Indicates if Wi-Fi AP is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WifiSTAEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_sta",
			Help:      "Indicates if Wi-Fi STA is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WifiRoamingThreshold: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "wifi_roaming_rssi_threshold",
			Help:      "RSSI threshold for Wi-Fi roaming",
		}, []string{"device_mac"}),
	}

	// Register all metrics with Prometheus
	prometheus.MustRegister(
		metrics.BLEEnabled,
		metrics.CloudEnabled,
		metrics.CloudServer,
		metrics.EthEnabled,
		metrics.EthIPv4Mode,
		metrics.WifiAPEnabled,
		metrics.WifiSTAEnabled,
		metrics.WifiRoamingThreshold,
	)
}

func UpdateShellyGetConfigMetrics(apiClient *client.APIClient) error {
	var config client.ShellyGetConfigResponse
	err := apiClient.FetchData("/rpc/Shelly.GetConfig", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config)

	return nil
}

// UpdateMetrics populates the metrics from the config structure
func (m *ConfigMetrics) UpdateMetrics(config client.ShellyGetConfigResponse) {
	// BLE
	if config.BLE.Enable {
		m.BLEEnabled.WithLabelValues(config.Sys.Device.MAC).Set(1)
	} else {
		m.BLEEnabled.WithLabelValues(config.Sys.Device.MAC).Set(0)
	}

	// Cloud
	if config.Cloud.Enable {
		m.CloudEnabled.WithLabelValues(config.Sys.Device.MAC).Set(1)
	} else {
		m.CloudEnabled.WithLabelValues(config.Sys.Device.MAC).Set(0)
	}
	m.CloudServer.WithLabelValues(config.Sys.Device.MAC, config.Cloud.Server).Set(1)

	// Ethernet
	if config.Eth.Enable {
		m.EthEnabled.WithLabelValues(config.Sys.Device.MAC).Set(1)
	} else {
		m.EthEnabled.WithLabelValues(config.Sys.Device.MAC).Set(0)
	}
	m.EthIPv4Mode.WithLabelValues(config.Sys.Device.MAC, config.Eth.IPv4Mode).Set(1)

	// Wi-Fi
	if config.Wifi.AP.Enable {
		m.WifiAPEnabled.WithLabelValues(config.Sys.Device.MAC).Set(1)
	} else {
		m.WifiAPEnabled.WithLabelValues(config.Sys.Device.MAC).Set(0)
	}
	if config.Wifi.STA.Enable {
		m.WifiSTAEnabled.WithLabelValues(config.Sys.Device.MAC).Set(1)
	} else {
		m.WifiSTAEnabled.WithLabelValues(config.Sys.Device.MAC).Set(0)
	}
	m.WifiRoamingThreshold.WithLabelValues(config.Sys.Device.MAC).Set(float64(config.Wifi.Roam.RSSIThreshold))
}
