package SwitchGetConfig

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type ConfigMetrics struct {
	InitialState         *prometheus.GaugeVec
	AutoOn               *prometheus.GaugeVec
	AutoOff              *prometheus.GaugeVec
	RecoverVoltageErrors *prometheus.GaugeVec
	PowerLimit           *prometheus.GaugeVec
	VoltageLimit         *prometheus.GaugeVec
	CurrentLimit         *prometheus.GaugeVec
}

var metrics *ConfigMetrics

func RegisterSwitchGetConfigMetrics() {
	metrics = &ConfigMetrics{
		InitialState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shely",
			Subsystem: "switch",
			Name:      "initial_state",
			Help:      "Initial state of the switch after power loss",
		}, []string{"device_mac", "switch_id"}),
		AutoOn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "auto_on",
			Help:      "Auto on behavior of switch",
		}, []string{"device_mac", "switch_id", "delay"}),
		AutoOff: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "auto_off",
			Help:      "Auto off behavior of switch",
		}, []string{"device_mac", "switch_id", "delay"}),
		RecoverVoltageErrors: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "recover_volate_errors",
			Help:      "Behavior of switch after voltage errors",
		}, []string{"device_mac", "switch_id"}),
		PowerLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "power_limit",
			Help:      "Power limit of switch in Watts",
		}, []string{"device_mac", "switch_id"}),
		VoltageLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "voltage_limit",
			Help:      "Voltage limits of the switch",
		}, []string{"device_mac", "switch_id", "kind"}),
		CurrentLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "current_limit",
			Help:      "Current limit in Amps",
		}, []string{"device_mac", "switch_id"}),
	}

	prometheus.MustRegister(
		metrics.InitialState,
		metrics.AutoOff,
		metrics.AutoOn,
		metrics.RecoverVoltageErrors,
		metrics.PowerLimit,
		metrics.VoltageLimit,
		metrics.CurrentLimit,
	)
}

func UpdateSwitchGetConfigMetrics(apiClient *client.APIClient, switchID int, device_mac string) error {
	var config client.SwitchGetConfigResponse
	err := apiClient.FetchData(fmt.Sprintf("/rpc/Switch.GetConfig?id=%d", switchID), &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config, device_mac)

	return nil
}

func (m *ConfigMetrics) UpdateMetrics(status client.SwitchGetConfigResponse, device_mac string) {
	switchID := fmt.Sprintf("%d", status.ID)

	switch state := status.InitialState; state {
	case "on":
		m.InitialState.WithLabelValues(device_mac, switchID).Set(1)
	case "off":
		m.InitialState.WithLabelValues(device_mac, switchID).Set(0)
	default:
		m.InitialState.WithLabelValues(device_mac, switchID).Set(2)
	}

	m.AutoOn.WithLabelValues(device_mac, switchID, fmt.Sprintf("%f", status.AutoOnDelay)).Set(boolToFloat64(status.AutoOn))
	m.AutoOff.WithLabelValues(device_mac, switchID, fmt.Sprintf("%f", status.AutoOffDelay)).Set(boolToFloat64(status.AutoOff))
	m.RecoverVoltageErrors.WithLabelValues(device_mac, switchID).Set(boolToFloat64(status.AutorecoverVoltageErrors))
	m.PowerLimit.WithLabelValues(device_mac, switchID).Set(float64(status.PowerLimit))
	m.VoltageLimit.WithLabelValues(device_mac, switchID, "overvoltage").Set(float64(status.VoltageLimit))
	m.VoltageLimit.WithLabelValues(device_mac, switchID, "undervoltage").Set(float64(status.UndervoltageLimit))
	m.CurrentLimit.WithLabelValues(device_mac, switchID).Set(status.CurrentLimit)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
