package ShellyGetStatus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type StatusMetrics struct {
	InputState        *prometheus.GaugeVec
	SwitchOutputState *prometheus.GaugeVec
	SwitchAPower      *prometheus.GaugeVec
	SwitchVoltage     *prometheus.GaugeVec
	SwitchCurrent     *prometheus.GaugeVec
	SwitchEnergy      *prometheus.GaugeVec
	SwitchTemperature *prometheus.GaugeVec
	SystemUptime      *prometheus.GaugeVec
	SystemRAMFree     *prometheus.GaugeVec
	SystemRAMSize     *prometheus.GaugeVec
	SystemFSFree      *prometheus.GaugeVec
	SystemFSSize      *prometheus.GaugeVec
	WIFIRSSI          *prometheus.GaugeVec
	UpdateAvailable   *prometheus.GaugeVec
}

var metrics *StatusMetrics

func RegisterShelly_GetStatusMetrics() {
	metrics = &StatusMetrics{
		InputState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "input_state",
			Help: "Indicates in which state an input is",
		}, []string{"device_mac", "input_id"}),
		SwitchOutputState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_state",
			Help: "State of switches (labels include switch ID)",
		}, []string{"device_mac", "switch_id"}),
		SwitchAPower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_apower",
			Help: "Apparent power of the switch in watts",
		}, []string{"device_mac", "switch_id"}),
		SwitchVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_voltage",
			Help: "Voltage of the switch in volts",
		}, []string{"device_mac", "switch_id"}),
		SwitchCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_current",
			Help: "Current of the switch in amperes",
		}, []string{"device_mac", "switch_id"}),
		SwitchEnergy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_energy",
			Help: "Energy consumption of the switch in kilowatt-hours",
		}, []string{"device_mac", "switch_id"}),
		SwitchTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "switch_temperature",
			Help: "Temperature of the switch in Celsius",
		}, []string{"device_mac", "switch_id"}),
		SystemUptime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_uptime",
			Help: "System uptime in seconds",
		}, []string{"device_mac"}),
		SystemRAMFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_ram_free",
			Help: "Amount of free RAM in bytes",
		}, []string{"device_mac"}),
		SystemRAMSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_ram_size",
			Help: "Total RAM in bytes",
		}, []string{"device_mac"}),
		SystemFSFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_fs_free",
			Help: "Amount of free FS in bytes",
		}, []string{"device_mac"}),
		SystemFSSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_fs_size",
			Help: "Total FS in bytes",
		}, []string{"device_mac"}),
		WIFIRSSI: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "wifi_rssi",
			Help: "Wi-Fi RSSI signal strength in dBm",
		}, []string{"device_mac", "ssid", "sta_ip"}),
		UpdateAvailable: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "update_available",
			Help: "Flag indicating whether an update is available (1 for available, 0 for not)",
		}, []string{"device_mac", "version"}),
	}

	prometheus.MustRegister(
		metrics.InputState,
		metrics.SwitchOutputState,
		metrics.SwitchAPower,
		metrics.SwitchVoltage,
		metrics.SwitchCurrent,
		metrics.SwitchEnergy,
		metrics.SwitchTemperature,
		metrics.SystemUptime,
		metrics.SystemRAMFree,
		metrics.SystemRAMSize,
		metrics.SystemFSFree,
		metrics.SystemFSSize,
		metrics.WIFIRSSI,
	)
}

func UpdateShelly_StatusMetrics(apiClient *client.APIClient) error {
	var config client.ShellyGetStatusResponse
	err := apiClient.FetchData("/rpc/Shelly.GetStatus", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config)

	return nil
}

func (m *StatusMetrics) UpdateMetrics(status client.ShellyGetStatusResponse) {
	deviceMAC := status.Sys.MAC

	for inputID, input := range status.Inputs {
		m.InputState.WithLabelValues(deviceMAC, inputID).Set(boolToFloat64(input.State))
	}

	for switchID, sw := range status.Switches {
		m.SwitchOutputState.WithLabelValues(deviceMAC, switchID).Set(boolToFloat64(sw.Output))
		m.SwitchAPower.WithLabelValues(deviceMAC, switchID).Set(sw.APower)
		m.SwitchVoltage.WithLabelValues(deviceMAC, switchID).Set(sw.Voltage)
		m.SwitchCurrent.WithLabelValues(deviceMAC, switchID).Set(sw.Current)
		m.SwitchEnergy.WithLabelValues(deviceMAC, switchID).Set(sw.AEnergy.Total)
		m.SwitchTemperature.WithLabelValues(deviceMAC, switchID).Set(sw.Temperature.TC)
	}

	m.SystemUptime.WithLabelValues(deviceMAC).Set(float64(status.Sys.Uptime))
	m.SystemRAMFree.WithLabelValues(deviceMAC).Set(float64(status.Sys.RAMFree))
	m.SystemRAMSize.WithLabelValues(deviceMAC).Set(float64(status.Sys.RAMSize))
	m.SystemFSFree.WithLabelValues(deviceMAC).Set(float64(status.Sys.FSFree))
	m.SystemFSSize.WithLabelValues(deviceMAC).Set(float64(status.Sys.FSSize))
	m.WIFIRSSI.WithLabelValues(deviceMAC, *status.Wifi.SSID, *status.Wifi.StaIP).Set(float64(status.Wifi.RSSI))
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
