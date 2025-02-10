package SwitchGetStatus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type SwitchGetStatusMetrics struct {
	State       *prometheus.GaugeVec
	APower      *prometheus.GaugeVec
	Voltage     *prometheus.GaugeVec
	Current     *prometheus.GaugeVec
	Freq        *prometheus.GaugeVec
	Energy      *prometheus.GaugeVec
	Temperature *prometheus.GaugeVec
}

var metrics *SwitchGetStatusMetrics

func RegisterSwitchGetStatusMetrics() {
	metrics = &SwitchGetStatusMetrics{
		State: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "state",
			Help:      "Describes the curren state the switch is in",
		}, []string{"device_mac", "switch_id"}),
		APower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "power",
			Help:      "Active power of the switch in Watts",
		}, []string{"device_mac", "switch_id"}),
		Voltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "voltage",
			Help:      "Present power in Volts",
		}, []string{"device_mac", "switch_id"}),
		Current: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "current",
			Help:      "Current draw by the switch in amps",
		}, []string{"device_mac", "switch_id"}),
		Freq: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "frequency",
			Help:      "Current input frequency of the power source in Hz.",
		}, []string{"device_mac", "switch_id"}),
		Energy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "energy",
			Help:      "Total consumption of the switch in Wh",
		}, []string{"device_mac", "switch_id"}),
		Temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "switch",
			Name:      "temperature",
			Help:      "Temerature of the shelly device in C or F",
		}, []string{"device_mac", "switch_id", "temperature_unit"}),
	}

	prometheus.MustRegister(
		metrics.State,
		metrics.APower,
		metrics.Voltage,
		metrics.Current,
		metrics.Freq,
		metrics.Energy,
		metrics.Temperature,
	)
}

func UpdateSwitchGetStatusMetrics(apiClient *client.APIClient, switchID int, device_mac string) error {
	var config client.SwitchGetStatusResponse
	err := apiClient.FetchData(fmt.Sprintf("/rpc/Switch.GetStatus?id=%d", switchID), &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config, device_mac)

	return nil
}

func (m *SwitchGetStatusMetrics) UpdateMetrics(status client.SwitchGetStatusResponse, device_mac string) {
	switchID := fmt.Sprintf("%d", status.ID)

	m.State.WithLabelValues(device_mac, switchID).Set(boolToFloat64(status.Output))
	m.APower.WithLabelValues(device_mac, switchID).Set(status.Apower)
	m.Voltage.WithLabelValues(device_mac, switchID).Set(status.Voltage)
	m.Current.WithLabelValues(device_mac, switchID).Set(status.Current)
	m.Freq.WithLabelValues(device_mac, switchID).Set(status.Freq)
	m.Energy.WithLabelValues(device_mac, switchID).Set(status.Aenergy.Total)
	m.Temperature.WithLabelValues(device_mac, switchID, "dC").Set(status.Temperature.TC)
	m.Temperature.WithLabelValues(device_mac, switchID, "dF").Set(status.Temperature.TF)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
