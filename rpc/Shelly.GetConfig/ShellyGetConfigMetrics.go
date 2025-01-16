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
	InputStates          *prometheus.GaugeVec
	SwitchAutoOnState    *prometheus.GaugeVec
	SwitchAutoOnDelays   *prometheus.GaugeVec
	SwitchAutoOffDelays  *prometheus.GaugeVec
	SwitchPowerLimits    *prometheus.GaugeVec
	WifiAPEnabled        *prometheus.GaugeVec
	WifiSTAEnabled       *prometheus.GaugeVec
	WifiRoamingThreshold *prometheus.GaugeVec
}

var metrics *ConfigMetrics

// RegisterMetrics initializes and registers the Prometheus metrics
func RegisterShelly_GetConfigMetrics() {
	metrics = &ConfigMetrics{
		BLEEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ble_enabled",
			Help: "Indicates if BLE is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		CloudEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloud_enabled",
			Help: "Indicates if Cloud is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		CloudServer: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cloud_server_info",
			Help: "Cloud server configuration (labels include server address)",
		}, []string{"device_mac", "server"}),
		EthEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eth_enabled",
			Help: "Indicates if Ethernet is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		EthIPv4Mode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eth_ipv4_mode",
			Help: "Ethernet IPv4 mode (labels include mode)",
		}, []string{"device_mac", "mode"}),
		InputStates: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "input_inverted",
			Help: "State of inputs (labels include input ID and type)",
		}, []string{"device_mac", "input_id", "type"}),
		SwitchAutoOnState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_auto_on",
			Help: "State of the automatic on feature (labels include switch ID)",
		}, []string{"device_mac", "switch_id"}),
		SwitchAutoOnDelays: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_auto_on_delay",
			Help: "Auto-on delay for switches (in seconds)",
		}, []string{"device_mac", "switch_id"}),
		SwitchAutoOffDelays: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_auto_off_delay",
			Help: "Auto-off delay for switches (in seconds)",
		}, []string{"device_mac", "switch_id"}),
		SwitchPowerLimits: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_power_limit",
			Help: "Power limit for switches (in watts)",
		}, []string{"device_mac", "switch_id"}),
		WifiAPEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wifi_ap_enabled",
			Help: "Indicates if Wi-Fi AP is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WifiSTAEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wifi_sta_enabled",
			Help: "Indicates if Wi-Fi STA is enabled (1 for true, 0 for false)",
		}, []string{"device_mac"}),
		WifiRoamingThreshold: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wifi_roaming_rssi_threshold",
			Help: "RSSI threshold for Wi-Fi roaming",
		}, []string{"device_mac"}),
	}

	// Register all metrics with Prometheus
	prometheus.MustRegister(
		metrics.BLEEnabled,
		metrics.CloudEnabled,
		metrics.CloudServer,
		metrics.EthEnabled,
		metrics.EthIPv4Mode,
		metrics.InputStates,
		metrics.SwitchAutoOnState,
		metrics.SwitchAutoOnDelays,
		metrics.SwitchAutoOffDelays,
		metrics.SwitchPowerLimits,
		metrics.WifiAPEnabled,
		metrics.WifiSTAEnabled,
		metrics.WifiRoamingThreshold,
	)
}

func UpdateShelly_GetConfigMetrics(apiClient *client.APIClient) error {
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

	// Inputs
	for id, input := range config.Inputs {
		state := 0.0
		if input.Invert {
			state = 1.0
		}
		m.InputStates.WithLabelValues(config.Sys.Device.MAC, id, input.Type).Set(state)
	}

	// Switches
	for id, sw := range config.Switches {
		state := 0.0
		if sw.AutoOn {
			state = 1.0
		}
		m.SwitchAutoOnState.WithLabelValues(config.Sys.Device.MAC, id).Set(state)
		m.SwitchAutoOnDelays.WithLabelValues(config.Sys.Device.MAC, id).Set(float64(sw.AutoOnDelay))
		m.SwitchAutoOffDelays.WithLabelValues(config.Sys.Device.MAC, id).Set(float64(sw.AutoOffDelay))
		m.SwitchPowerLimits.WithLabelValues(config.Sys.Device.MAC, id).Set(float64(sw.PowerLimit))
	}

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
