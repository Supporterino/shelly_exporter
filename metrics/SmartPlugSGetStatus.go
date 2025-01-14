package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

var (
	// Cloud Metrics
	cloudConnectedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_cloud_connected",
		Help: "Indicates if the device is connected to the cloud.",
	}, []string{"device_mac"})

	// MQTT Metrics
	mqttConnectedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_mqtt_connected",
		Help: "Indicates if the device is connected to the MQTT broker.",
	}, []string{"device_mac"})

	// Switch Metrics
	switchOutputGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_output",
		Help: "Indicates if the switch output is active.",
	}, []string{"device_mac", "switch_id"})
	switchAPowerGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_apower",
		Help: "Active power in watts.",
	}, []string{"device_mac", "switch_id"})
	switchVoltageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_voltage",
		Help: "Voltage in volts.",
	}, []string{"device_mac", "switch_id"})
	switchCurrentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_current",
		Help: "Current in amperes.",
	}, []string{"device_mac", "switch_id"})
	switchEnergyTotalGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_aenergy_total",
		Help: "Total accumulated energy in kWh.",
	}, []string{"device_mac", "switch_id"})
	switchTemperatureCGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_temperature_c",
		Help: "Switch temperature in Celsius.",
	}, []string{"device_mac", "switch_id"})

	// System Metrics
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

	// WiFi Metrics
	wifiRSSIGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_wifi_rssi",
		Help: "WiFi signal strength in dBm.",
	}, []string{"device_mac", "wifi_ssid"})
)

func fetchAndUpdateStatusMetrics(apiClient *client.APIClient) error {
	var status client.ShellyGetStatusResponse
	err := apiClient.FetchData("/rpc/Shelly.GetStatus", &status)
	if err != nil {
		return fmt.Errorf("error fetching status: %w", err)
	}

	updateStatusMetrics(status)
	return nil
}

func updateStatusMetrics(status client.ShellyGetStatusResponse) {
	labels := map[string]string{
		"device_mac": status.Sys.MAC,
	}

	// Cloud Metrics
	cloudConnectedGauge.With(labels).Set(boolToFloat64(status.Cloud.Connected))

	// MQTT Metrics
	mqttConnectedGauge.With(labels).Set(boolToFloat64(status.MQTT.Connected))

	// Switch Metrics
	switchLabels := map[string]string{
		"device_mac": status.Sys.MAC,
		"switch_id":  "switch:0",
	}
	switchOutputGauge.With(switchLabels).Set(boolToFloat64(status.Switch0.Output))
	switchAPowerGauge.With(switchLabels).Set(status.Switch0.APower)
	switchVoltageGauge.With(switchLabels).Set(status.Switch0.Voltage)
	switchCurrentGauge.With(switchLabels).Set(status.Switch0.Current)
	switchEnergyTotalGauge.With(switchLabels).Set(status.Switch0.AEnergy.Total)
	switchTemperatureCGauge.With(switchLabels).Set(status.Switch0.Temperature.TC)

	// System Metrics
	sysUptimeGauge.With(labels).Set(float64(status.Sys.Uptime))
	sysRAMFreeGauge.With(labels).Set(float64(status.Sys.RAMFree))
	sysFSFreeGauge.With(labels).Set(float64(status.Sys.FSFree))

	// WiFi Metrics
	wifiLabels := map[string]string{
		"device_mac": status.Sys.MAC,
		"wifi_ssid":  status.WiFi.SSID,
	}
	wifiRSSIGauge.With(wifiLabels).Set(float64(status.WiFi.RSSI))
}
