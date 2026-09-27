package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// SwitchStatusCollector collects switch metrics from Switch.GetStatus.
type SwitchStatusCollector struct {
	State       *prometheus.GaugeVec
	APower      *prometheus.GaugeVec
	Voltage     *prometheus.GaugeVec
	Current     *prometheus.GaugeVec
	Freq        *prometheus.GaugeVec
	Energy      *prometheus.GaugeVec
	Temperature *prometheus.GaugeVec
}

// NewSwitchStatusCollector constructs and registers a SwitchStatusCollector.
func NewSwitchStatusCollector(reg prometheus.Registerer) (*SwitchStatusCollector, error) {
	c := &SwitchStatusCollector{
		State: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "state",
			Help:      "Describes the curren state the switch is in",
		}, []string{"device_mac", "switch_id", "name"}),
		APower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "power",
			Help:      "Active power of the switch in Watts",
		}, []string{"device_mac", "switch_id", "name"}),
		Voltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "voltage",
			Help:      "Present power in Volts",
		}, []string{"device_mac", "switch_id", "name"}),
		Current: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "current",
			Help:      "Current draw by the switch in amps",
		}, []string{"device_mac", "switch_id", "name"}),
		Freq: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "frequency",
			Help:      "Current input frequency of the power source in Hz.",
		}, []string{"device_mac", "switch_id", "name"}),
		Energy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "energy",
			Help:      "Total consumption of the switch in Wh",
		}, []string{"device_mac", "switch_id", "name"}),
		Temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "temperature",
			Help:      "Temerature of the shelly device in C or F",
		}, []string{"device_mac", "switch_id", "name", "temperature_unit"}),
	}

	if err := register(reg,
		c.State,
		c.APower,
		c.Voltage,
		c.Current,
		c.Freq,
		c.Energy,
		c.Temperature,
	); err != nil {
		return nil, fmt.Errorf("failed to register switch status collector: %w", err)
	}

	return c, nil
}

// Update fetches Switch.GetStatus and updates the collector metrics. name is the
// human-readable channel name reported in the device configuration.
func (c *SwitchStatusCollector) Update(ctx context.Context, fetcher client.Fetcher, switchID int, name, deviceMAC string) error {
	var status client.SwitchGetStatusResponse
	if err := fetcher.FetchData(ctx, fmt.Sprintf("/rpc/Switch.GetStatus?id=%d", switchID), &status); err != nil {
		return fmt.Errorf("failed to fetch Switch.GetStatus: %w", err)
	}

	c.UpdateMetrics(status, deviceMAC, name)
	return nil
}

// UpdateMetrics populates the metrics from a Switch.GetStatus response.
func (c *SwitchStatusCollector) UpdateMetrics(status client.SwitchGetStatusResponse, deviceMAC, name string) {
	switchID := fmt.Sprintf("%d", status.ID)

	c.State.WithLabelValues(deviceMAC, switchID, name).Set(boolToFloat64(status.Output))
	setGauge(c.APower, status.Apower, deviceMAC, switchID, name)
	setGauge(c.Voltage, status.Voltage, deviceMAC, switchID, name)
	setGauge(c.Current, status.Current, deviceMAC, switchID, name)
	setGauge(c.Freq, status.Freq, deviceMAC, switchID, name)
	if status.Aenergy != nil {
		c.Energy.WithLabelValues(deviceMAC, switchID, name).Set(status.Aenergy.Total)
	}
	if status.Temperature != nil {
		c.Temperature.WithLabelValues(deviceMAC, switchID, name, "dC").Set(status.Temperature.TC)
		c.Temperature.WithLabelValues(deviceMAC, switchID, name, "dF").Set(status.Temperature.TF)
	}
}
