package ShellySmartPlugS

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

var (
	// Inputs
	inputStateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_input_state",
		Help: "State of each input (true=1, false=0).",
	}, []string{"device_mac", "input_id"})

	// Switches
	switchOutputGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_output",
		Help: "Output state of each switch (true=1, false=0).",
	}, []string{"device_mac", "switch_id"})
	switchAPowerGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_apower",
		Help: "Active power of each switch in watts.",
	}, []string{"device_mac", "switch_id"})
	switchVoltageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_voltage",
		Help: "Voltage of each switch in volts.",
	}, []string{"device_mac", "switch_id"})
	switchAEnergyTotalGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_aenergy_total",
		Help: "Total accumulated energy of each switch in kWh.",
	}, []string{"device_mac", "switch_id"})
	switchTemperatureCGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_temperature_c",
		Help: "Temperature of each switch in Celsius.",
	}, []string{"device_mac", "switch_id"})

	// System
	sysUptimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_sys_uptime_seconds",
		Help: "Uptime of the device in seconds.",
	}, []string{"device_mac"})
	sysRAMFreeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_sys_ram_free",
		Help: "Free RAM in bytes.",
	}, []string{"device_mac"})
	sysFSFreeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_sys_fs_free",
		Help: "Free filesystem space in bytes.",
	}, []string{"device_mac"})

	// WiFi
	wifiRSSIGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_wifi_rssi",
		Help: "WiFi signal strength in dBm.",
	}, []string{"device_mac", "wifi_ssid"})

	// Ethernet
	ethIPGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_eth_ip",
		Help: "Exposes the Ethernet IP address as a label.",
	}, []string{"device_mac", "eth_ip"})
)

func fetchAndUpdateStatusMetrics(apiClient *client.APIClient) error {
	status, err := fetchAndProcessStatus(apiClient)

	if err != nil {
		return fmt.Errorf("failed to fetch status: %w", err)
	}

	updateStatusMetrics(*status)
	return nil
}

// fetchAndProcessStatus fetches the status and processes dynamic keys.
func fetchAndProcessStatus(apiClient *client.APIClient) (*client.ShellyGetStatusResponse, error) {
	var status client.ShellyGetStatusResponse
	err := apiClient.FetchData("/rpc/Shelly.GetStatus", &status)

	if err != nil {
		return nil, fmt.Errorf("error fetching status: %w", err)
	}

	return &status, nil
}

func updateStatusMetrics(status client.ShellyGetStatusResponse) {
	deviceMAC := status.Sys.MAC

	// Inputs
	for inputID, input := range status.Inputs {
		inputStateGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"input_id":   inputID,
		}).Set(boolToFloat64(input.State))
	}

	// Switches
	for switchID, sw := range status.Switches {
		switchOutputGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"switch_id":  switchID,
		}).Set(boolToFloat64(sw.Output))

		switchAPowerGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"switch_id":  switchID,
		}).Set(sw.APower)

		switchVoltageGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"switch_id":  switchID,
		}).Set(sw.Voltage)

		switchAEnergyTotalGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"switch_id":  switchID,
		}).Set(sw.AEnergy.Total)

		switchTemperatureCGauge.With(prometheus.Labels{
			"device_mac": deviceMAC,
			"switch_id":  switchID,
		}).Set(sw.Temperature.TC)
	}

	// System
	sysUptimeGauge.With(prometheus.Labels{
		"device_mac": deviceMAC,
	}).Set(float64(status.Sys.Uptime))

	sysRAMFreeGauge.With(prometheus.Labels{
		"device_mac": deviceMAC,
	}).Set(float64(status.Sys.RAMFree))

	sysFSFreeGauge.With(prometheus.Labels{
		"device_mac": deviceMAC,
	}).Set(float64(status.Sys.FSFree))

	// WiFi
	wifiRSSIGauge.With(prometheus.Labels{
		"device_mac": deviceMAC,
		"wifi_ssid":  coalesce(status.Wifi.SSID, "unknown"),
	}).Set(float64(status.Wifi.RSSI))

	// Ethernet
	ethIPGauge.With(prometheus.Labels{
		"device_mac": deviceMAC,
		"eth_ip":     coalesce(&status.Eth.IP, "unknown"),
	}).Set(1)
}
