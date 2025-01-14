package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

var (
	apiCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests made.",
		},
	)

	// BLE Metrics
	bleEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_ble_enabled",
		Help: "Indicates if BLE is enabled.",
	})
	bleRPCEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_ble_rpc_enabled",
		Help: "Indicates if BLE RPC is enabled.",
	})
	bleObserverEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_ble_observer_enabled",
		Help: "Indicates if BLE Observer is enabled.",
	})

	// Cloud Metrics
	cloudEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_cloud_enabled",
		Help: "Indicates if cloud is enabled.",
	})

	// MQTT Metrics
	mqttEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_mqtt_enabled",
		Help: "Indicates if MQTT is enabled.",
	})
	mqttRPCNotificationsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_mqtt_rpc_notifications",
		Help: "Indicates if MQTT RPC notifications are enabled.",
	})
	mqttStatusNotificationsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_mqtt_status_notifications",
		Help: "Indicates if MQTT status notifications are enabled.",
	})

	// Switch Metrics
	switchAutoOnGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_switch_auto_on",
		Help: "Indicates if auto-on is enabled for the switch.",
	})
	switchAutoOnDelayGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_switch_auto_on_delay",
		Help: "Auto-on delay for the switch in seconds.",
	})
	switchPowerLimitGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_switch_power_limit",
		Help: "Power limit of the switch in watts.",
	})
	switchVoltageLimitGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_switch_voltage_limit",
		Help: "Voltage limit of the switch in volts.",
	})
	switchCurrentLimitGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_switch_current_limit",
		Help: "Current limit of the switch in amperes.",
	})

	// WiFi Metrics
	wifiAPEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_wifi_ap_enabled",
		Help: "Indicates if WiFi AP mode is enabled.",
	})
	wifiSTAEnabledGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_wifi_sta_enabled",
		Help: "Indicates if WiFi STA mode is enabled.",
	})
	wifiRSSIThresholdGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "shelly_wifi_rssi_threshold",
		Help: "RSSI threshold for WiFi roaming.",
	})
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register() {
	// Register all metrics with Prometheus
	prometheus.MustRegister(
		apiCounter,
		bleEnabledGauge,
		bleRPCEnabledGauge,
		bleObserverEnabledGauge,
		cloudEnabledGauge,
		mqttEnabledGauge,
		mqttRPCNotificationsGauge,
		mqttStatusNotificationsGauge,
		switchAutoOnGauge,
		switchAutoOnDelayGauge,
		switchPowerLimitGauge,
		switchVoltageLimitGauge,
		switchCurrentLimitGauge,
		wifiAPEnabledGauge,
		wifiSTAEnabledGauge,
		wifiRSSIThresholdGauge,
	)

	apiClient := client.NewAPIClient("http://10.1.255.111", 10*time.Second)

	// Start fetching metrics periodically
	go func() {
		for {
			err := fetchAndUpdateMetrics(apiClient)
			if err != nil {
				log.Printf("Error fetching metrics: %v", err)
			}
			time.Sleep(30 * time.Second) // Adjust interval as needed
		}
	}()
}

// fetchAndUpdateMetrics fetches data from the API and updates Prometheus metrics.
func fetchAndUpdateMetrics(apiClient *client.APIClient) error {
	err := fetchAndUpdateConfigMetrics(apiClient)
	if err != nil {
		return fmt.Errorf("failed to update config metrics: %w", err)
	}

	return nil
}

func fetchAndUpdateConfigMetrics(apiClient *client.APIClient) error {
	apiCounter.Inc()

	var config client.ShellyGetConfigResponse
	err := apiClient.FetchData("/rpc/Shelly.GetConfig", &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	updateMetrics(config)
	return nil
}

func updateMetrics(config client.ShellyGetConfigResponse) {
	// Update BLE Metrics
	bleEnabledGauge.Set(boolToFloat64(config.BLE.Enable))
	bleRPCEnabledGauge.Set(boolToFloat64(config.BLE.RPC.Enable))
	bleObserverEnabledGauge.Set(boolToFloat64(config.BLE.Observer.Enable))

	// Update Cloud Metrics
	cloudEnabledGauge.Set(boolToFloat64(config.Cloud.Enable))

	// Update MQTT Metrics
	mqttEnabledGauge.Set(boolToFloat64(config.MQTT.Enable))
	mqttRPCNotificationsGauge.Set(boolToFloat64(config.MQTT.RPCNotifications))
	mqttStatusNotificationsGauge.Set(boolToFloat64(config.MQTT.StatusNotifications))

	// Update Switch Metrics
	switchAutoOnGauge.Set(boolToFloat64(config.Switch0.AutoOn))
	switchAutoOnDelayGauge.Set(config.Switch0.AutoOnDelay)
	switchPowerLimitGauge.Set(float64(config.Switch0.PowerLimit))
	switchVoltageLimitGauge.Set(float64(config.Switch0.VoltageLimit))
	switchCurrentLimitGauge.Set(config.Switch0.CurrentLimit)

	// Update WiFi Metrics
	wifiAPEnabledGauge.Set(boolToFloat64(config.WiFi.AP.Enable))
	wifiSTAEnabledGauge.Set(boolToFloat64(config.WiFi.STA.Enable))
	wifiRSSIThresholdGauge.Set(float64(config.WiFi.Roam.RSSIThreshold))
}

// boolToFloat64 converts a boolean value to float64 (1 for true, 0 for false).
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
