package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// SwitchConfigCollector collects switch configuration metrics from Switch.GetConfig.
type SwitchConfigCollector struct {
	InitialState         *prometheus.GaugeVec
	AutoOn               *prometheus.GaugeVec
	AutoOff              *prometheus.GaugeVec
	RecoverVoltageErrors *prometheus.GaugeVec
	PowerLimit           *prometheus.GaugeVec
	VoltageLimit         *prometheus.GaugeVec
	CurrentLimit         *prometheus.GaugeVec
}

// NewSwitchConfigCollector constructs and registers a SwitchConfigCollector.
func NewSwitchConfigCollector(reg prometheus.Registerer) (*SwitchConfigCollector, error) {
	c := &SwitchConfigCollector{
		InitialState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "initial_state",
			Help:      "Initial state of the switch after power loss",
		}, []string{"device_mac", "switch_id", "name"}),
		AutoOn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "auto_on",
			Help:      "Auto on behavior of switch",
		}, []string{"device_mac", "switch_id", "name", "delay"}),
		AutoOff: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "auto_off",
			Help:      "Auto off behavior of switch",
		}, []string{"device_mac", "switch_id", "name", "delay"}),
		RecoverVoltageErrors: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "recover_volate_errors",
			Help:      "Behavior of switch after voltage errors",
		}, []string{"device_mac", "switch_id", "name"}),
		PowerLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "power_limit",
			Help:      "Power limit of switch in Watts",
		}, []string{"device_mac", "switch_id", "name"}),
		VoltageLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "voltage_limit",
			Help:      "Voltage limits of the switch",
		}, []string{"device_mac", "switch_id", "name", "kind"}),
		CurrentLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "current_limit",
			Help:      "Current limit in Amps",
		}, []string{"device_mac", "switch_id", "name"}),
	}

	if err := register(reg,
		c.InitialState,
		c.AutoOn,
		c.AutoOff,
		c.RecoverVoltageErrors,
		c.PowerLimit,
		c.VoltageLimit,
		c.CurrentLimit,
	); err != nil {
		return nil, fmt.Errorf("failed to register switch config collector: %w", err)
	}

	return c, nil
}

// Update fetches Switch.GetConfig and updates the collector metrics.
func (c *SwitchConfigCollector) Update(ctx context.Context, fetcher client.Fetcher, switchID int, deviceMAC string) error {
	var cfg client.SwitchGetConfigResponse
	if err := fetcher.FetchData(ctx, fmt.Sprintf("/rpc/Switch.GetConfig?id=%d", switchID), &cfg); err != nil {
		return fmt.Errorf("failed to fetch Switch.GetConfig: %w", err)
	}

	c.UpdateMetrics(cfg, deviceMAC)
	return nil
}

// UpdateMetrics populates the metrics from a Switch.GetConfig response.
func (c *SwitchConfigCollector) UpdateMetrics(cfg client.SwitchGetConfigResponse, deviceMAC string) {
	switchID := fmt.Sprintf("%d", cfg.ID)
	name := cfg.Name

	switch cfg.InitialState {
	case "on":
		c.InitialState.WithLabelValues(deviceMAC, switchID, name).Set(1)
	case "off":
		c.InitialState.WithLabelValues(deviceMAC, switchID, name).Set(0)
	case "restore_last":
		c.InitialState.WithLabelValues(deviceMAC, switchID, name).Set(2)
	case "match_input":
		c.InitialState.WithLabelValues(deviceMAC, switchID, name).Set(3)
	default:
		c.InitialState.WithLabelValues(deviceMAC, switchID, name).Set(-1)
	}

	if cfg.AutoOn != nil && cfg.AutoOnDelay != nil {
		c.AutoOn.WithLabelValues(deviceMAC, switchID, name, fmt.Sprintf("%f", *cfg.AutoOnDelay)).Set(boolToFloat64(*cfg.AutoOn))
	}
	if cfg.AutoOff != nil && cfg.AutoOffDelay != nil {
		c.AutoOff.WithLabelValues(deviceMAC, switchID, name, fmt.Sprintf("%f", *cfg.AutoOffDelay)).Set(boolToFloat64(*cfg.AutoOff))
	}
	if cfg.AutorecoverVoltageErrors != nil {
		c.RecoverVoltageErrors.WithLabelValues(deviceMAC, switchID, name).Set(boolToFloat64(*cfg.AutorecoverVoltageErrors))
	}
	setGauge(c.PowerLimit, cfg.PowerLimit, deviceMAC, switchID, name)
	if cfg.VoltageLimit != nil {
		c.VoltageLimit.WithLabelValues(deviceMAC, switchID, name, "overvoltage").Set(*cfg.VoltageLimit)
	}
	if cfg.UndervoltageLimit != nil {
		c.VoltageLimit.WithLabelValues(deviceMAC, switchID, name, "undervoltage").Set(*cfg.UndervoltageLimit)
	}
	setGauge(c.CurrentLimit, cfg.CurrentLimit, deviceMAC, switchID, name)
}
