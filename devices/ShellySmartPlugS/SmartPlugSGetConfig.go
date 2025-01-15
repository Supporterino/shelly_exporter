package ShellySmartPlugS

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

var (
	apiCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests made.",
		},
		[]string{"device_name", "device_mac"}, // Labels for device context
	)

	// BLE Metrics
	bleEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_ble_enabled",
		Help: "Indicates if BLE is enabled.",
	}, []string{"device_name", "device_mac"})

	bleRPCEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_ble_rpc_enabled",
		Help: "Indicates if BLE RPC is enabled.",
	}, []string{"device_name", "device_mac"})

	bleObserverEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_ble_observer_enabled",
		Help: "Indicates if BLE Observer is enabled.",
	}, []string{"device_name", "device_mac"})

	// MQTT Metrics
	mqttEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_mqtt_enabled",
		Help: "Indicates if MQTT is enabled.",
	}, []string{"device_name", "device_mac", "server"})
	mqttRPCNotificationsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_mqtt_rpc_notifications",
		Help: "Indicates if MQTT RPC notifications are enabled.",
	}, []string{"device_name", "device_mac", "server"})
	mqttStatusNotificationsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_mqtt_status_notifications",
		Help: "Indicates if MQTT status notifications are enabled.",
	}, []string{"device_name", "device_mac", "server"})

	// Cloud Metrics
	cloudEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_cloud_enabled",
		Help: "Indicates if cloud is enabled.",
	}, []string{"device_name", "device_mac"})

	// Switch Metrics
	switchAutoOnGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_auto_on",
		Help: "Indicates if auto-on is enabled for the switch.",
	}, []string{"device_name", "device_mac", "switch_id", "switch_name"})

	switchPowerLimitGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_power_limit",
		Help: "Power limit of the switch in watts.",
	}, []string{"device_name", "device_mac", "switch_id", "switch_name"})

	switchAutoOnDelayGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_auto_on_delay",
		Help: "Auto-on delay for the switch in seconds.",
	}, []string{"device_name", "device_mac", "switch_id", "switch_name"})
	switchVoltageLimitGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_voltage_limit",
		Help: "Voltage limit of the switch in volts.",
	}, []string{"device_name", "device_mac", "switch_id", "switch_name"})
	switchCurrentLimitGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_switch_current_limit",
		Help: "Current limit of the switch in amperes.",
	}, []string{"device_name", "device_mac", "switch_id", "switch_name"})

	// WiFi Metrics
	wifiAPEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_wifi_ap_enabled",
		Help: "Indicates if WiFi AP mode is enabled.",
	}, []string{"device_name", "device_mac", "wifi_ssid"})

	wifiSTAEnabledGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_wifi_sta_enabled",
		Help: "Indicates if WiFi STA mode is enabled.",
	}, []string{"device_name", "device_mac", "wifi_ssid"})
	wifiRSSIThresholdGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shelly_wifi_rssi_threshold",
		Help: "RSSI threshold for WiFi roaming.",
	}, []string{"device_name", "device_mac", "wifi_ssid"})
)

func fetchAndUpdateConfigMetrics(apiClient *client.APIClient) error {
	var config client.ShellyGetConfigResponse
	err := apiClient.FetchData("/rpc/Shelly.GetConfig", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	updateConfigMetrics(config)
	return nil
}

func updateConfigMetrics(config client.ShellyGetConfigResponse) {
	labels := map[string]string{
		"device_name": config.Sys.Device.Name,
		"device_mac":  config.Sys.Device.MAC,
	}

	// Update API Counter
	apiCounter.With(labels).Inc()

	// Update BLE Metrics
	bleEnabledGauge.With(labels).Set(boolToFloat64(config.BLE.Enable))
	bleRPCEnabledGauge.With(labels).Set(boolToFloat64(config.BLE.RPC.Enable))
	bleObserverEnabledGauge.With(labels).Set(boolToFloat64(config.BLE.Observer.Enable))

	// Update Cloud Metrics
	cloudEnabledGauge.With(labels).Set(boolToFloat64(config.Cloud.Enable))

	// Update MQTT Metrics
	mqttLabels := map[string]string{
		"device_name": config.Sys.Device.Name,
		"device_mac":  config.Sys.Device.MAC,
		"server":      config.MQTT.Server,
	}

	mqttEnabledGauge.With(mqttLabels).Set(boolToFloat64(config.MQTT.Enable))
	mqttRPCNotificationsGauge.With(mqttLabels).Set(boolToFloat64(config.MQTT.RPCNotifications))
	mqttStatusNotificationsGauge.With(mqttLabels).Set(boolToFloat64(config.MQTT.StatusNotifications))

	// Update Switch Metrics
	switchLabels := map[string]string{
		"device_name": config.Sys.Device.Name,
		"device_mac":  config.Sys.Device.MAC,
		"switch_id":   "switch:0",
		"switch_name": fmt.Sprintf("%v", config.Switch0.Name), // Convert nil to string if necessary
	}
	switchAutoOnGauge.With(switchLabels).Set(boolToFloat64(config.Switch0.AutoOn))
	switchPowerLimitGauge.With(switchLabels).Set(float64(config.Switch0.PowerLimit))
	switchAutoOnGauge.With(switchLabels).Set(boolToFloat64(config.Switch0.AutoOn))
	switchAutoOnDelayGauge.With(switchLabels).Set(config.Switch0.AutoOnDelay)
	switchVoltageLimitGauge.With(switchLabels).Set(float64(config.Switch0.VoltageLimit))
	switchCurrentLimitGauge.With(switchLabels).Set(config.Switch0.CurrentLimit)

	// Update WiFi Metrics
	wifiLabels := map[string]string{
		"device_name": config.Sys.Device.Name,
		"device_mac":  config.Sys.Device.MAC,
		"wifi_ssid":   config.WiFi.STA.SSID,
	}
	wifiAPEnabledGauge.With(wifiLabels).Set(boolToFloat64(config.WiFi.AP.Enable))
	wifiSTAEnabledGauge.With(wifiLabels).Set(boolToFloat64(config.WiFi.STA.Enable))
	wifiRSSIThresholdGauge.With(wifiLabels).Set(float64(config.WiFi.Roam.RSSIThreshold))
}
