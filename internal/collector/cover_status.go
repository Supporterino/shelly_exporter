package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// CoverStatusCollector collects cover metrics from Cover.GetStatus.
type CoverStatusCollector struct {
	State       *prometheus.GaugeVec
	APower      *prometheus.GaugeVec
	Voltage     *prometheus.GaugeVec
	Current     *prometheus.GaugeVec
	Pf          *prometheus.GaugeVec
	Freq        *prometheus.GaugeVec
	Energy      *prometheus.GaugeVec
	Temperature *prometheus.GaugeVec
	PosControl  *prometheus.GaugeVec
	Position    *prometheus.GaugeVec
}

// NewCoverStatusCollector constructs and registers a CoverStatusCollector.
func NewCoverStatusCollector(reg prometheus.Registerer) (*CoverStatusCollector, error) {
	c := &CoverStatusCollector{
		State: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "state",
			Help:      "Describes the current postion aka state the cover is in. (1 = open, 0 = closed, 2 = in movenment, 3 = stopped)",
		}, []string{"device_mac", "cover_id", "name"}),
		APower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "power",
			Help:      "Active power of the cover in Watts",
		}, []string{"device_mac", "cover_id", "name"}),
		Voltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "voltage",
			Help:      "Present power in Volts",
		}, []string{"device_mac", "cover_id", "name"}),
		Current: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "current",
			Help:      "Current draw by the cover in amps",
		}, []string{"device_mac", "cover_id", "name"}),
		Pf: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "powerfactor",
			Help:      "Power factor of the cover",
		}, []string{"device_mac", "cover_id", "name"}),
		Freq: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "frequency",
			Help:      "Current input frequency of the power source in Hz.",
		}, []string{"device_mac", "cover_id", "name"}),
		Energy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "energy",
			Help:      "Total consumption of the cover in Wh",
		}, []string{"device_mac", "cover_id", "name"}),
		Temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "temperature",
			Help:      "Temerature of the shelly device in C or F",
		}, []string{"device_mac", "cover_id", "name", "temperature_unit"}),
		PosControl: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "pos_control",
			Help:      "Boolean indicating if position control is present",
		}, []string{"device_mac", "cover_id", "name"}),
		Position: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "position",
			Help:      "Current position of the cover",
		}, []string{"device_mac", "cover_id", "name"}),
	}

	if err := register(reg,
		c.State,
		c.APower,
		c.Voltage,
		c.Current,
		c.Pf,
		c.Freq,
		c.Energy,
		c.Temperature,
		c.PosControl,
		c.Position,
	); err != nil {
		return nil, fmt.Errorf("failed to register cover status collector: %w", err)
	}

	return c, nil
}

// Update fetches Cover.GetStatus and updates the collector metrics. name is the
// human-readable channel name reported in the device configuration.
func (c *CoverStatusCollector) Update(ctx context.Context, fetcher client.Fetcher, coverID int, name, deviceMAC string) error {
	var status client.CoverGetStatusResponse
	if err := fetcher.FetchData(ctx, fmt.Sprintf("/rpc/Cover.GetStatus?id=%d", coverID), &status); err != nil {
		return fmt.Errorf("failed to fetch Cover.GetStatus: %w", err)
	}

	c.UpdateMetrics(status, deviceMAC, name)
	return nil
}

// UpdateMetrics populates the metrics from a Cover.GetStatus response.
func (c *CoverStatusCollector) UpdateMetrics(status client.CoverGetStatusResponse, deviceMAC, name string) {
	coverID := fmt.Sprintf("%d", status.ID)

	switch status.State {
	case "open":
		c.State.WithLabelValues(deviceMAC, coverID, name).Set(1)
	case "closed":
		c.State.WithLabelValues(deviceMAC, coverID, name).Set(0)
	case "opening", "closing", "calibrating":
		c.State.WithLabelValues(deviceMAC, coverID, name).Set(2)
	case "stopped":
		c.State.WithLabelValues(deviceMAC, coverID, name).Set(3)
	default:
		c.State.WithLabelValues(deviceMAC, coverID, name).Set(-1)
	}

	setGauge(c.APower, status.Apower, deviceMAC, coverID, name)
	setGauge(c.Voltage, status.Voltage, deviceMAC, coverID, name)
	setGauge(c.Current, status.Current, deviceMAC, coverID, name)
	setGauge(c.Pf, status.Pf, deviceMAC, coverID, name)
	setGauge(c.Freq, status.Freq, deviceMAC, coverID, name)
	if status.Aenergy != nil {
		c.Energy.WithLabelValues(deviceMAC, coverID, name).Set(status.Aenergy.Total)
	}
	if status.Temperature != nil {
		c.Temperature.WithLabelValues(deviceMAC, coverID, name, "dC").Set(status.Temperature.TC)
		c.Temperature.WithLabelValues(deviceMAC, coverID, name, "dF").Set(status.Temperature.TF)
	}
	if status.PosControl != nil {
		c.PosControl.WithLabelValues(deviceMAC, coverID, name).Set(boolToFloat64(*status.PosControl))
	}
	if status.CurrentPos != nil {
		c.Position.WithLabelValues(deviceMAC, coverID, name).Set(float64(*status.CurrentPos))
	}
}
