package CoverGetStatus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type StatusMetrics struct {
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

var metrics *StatusMetrics

func RegisterCoverGetStatusMetrics() {
	metrics = &StatusMetrics{
		State: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "state",
			Help:      "Describes the current postion aka state the cover is in. (1 = open, 0 = closed, 2 = in movenment, 3 = stopped)",
		}, []string{"device_mac", "cover_id"}),
		APower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "power",
			Help:      "Apparent power of the cover in Watts",
		}, []string{"device_mac", "cover_id"}),
		Voltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "voltage",
			Help:      "Present power in Volts",
		}, []string{"device_mac", "cover_id"}),
		Current: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "current",
			Help:      "Current draw by the cover in amps",
		}, []string{"device_mac", "cover_id"}),
		Pf: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "powerfactor",
			Help:      "Power factor of the cover",
		}, []string{"device_mac", "cover_id"}),
		Freq: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "frequency",
			Help:      "Current input frequency of the power source in Hz.",
		}, []string{"device_mac", "cover_id"}),
		Energy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "energy",
			Help:      "Total consumption of the cover in Wh",
		}, []string{"device_mac", "cover_id"}),
		Temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "temperature",
			Help:      "Temerature of the shelly device in C or F",
		}, []string{"device_mac", "cover_id", "temperature_unit"}),
		PosControl: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "pos_control",
			Help:      "Boolean indicating if position control is present",
		}, []string{"device_mac", "cover_id"}),
		Position: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "cover",
			Name:      "position",
			Help:      "Current position of the cover",
		}, []string{"device_mac", "cover_id"}),
	}

	prometheus.MustRegister(
		metrics.State,
		metrics.APower,
		metrics.Voltage,
		metrics.Current,
		metrics.Pf,
		metrics.Freq,
		metrics.Energy,
		metrics.Temperature,
		metrics.PosControl,
		metrics.Position,
	)
}

func UpdateCoverGetStatusMetrics(apiClient *client.APIClient, coverID int, device_mac string) error {
	var config client.CoverGetStatusResponse
	err := apiClient.FetchData(fmt.Sprintf("/rpc/Cover.GetStatus?id=%d", coverID), &config)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(config, device_mac)

	return nil
}

func (m *StatusMetrics) UpdateMetrics(status client.CoverGetStatusResponse, device_mac string) {
	coverID := fmt.Sprintf("%d", status.ID)

	switch state := status.State; state {
	case "open":
		m.State.WithLabelValues(device_mac, coverID).Set(1)
	case "closed":
		m.State.WithLabelValues(device_mac, coverID).Set(0)
	case "opening":
		m.State.WithLabelValues(device_mac, coverID).Set(2)
	case "closing":
		m.State.WithLabelValues(device_mac, coverID).Set(2)
	case "stopped":
		m.State.WithLabelValues(device_mac, coverID).Set(3)
	case "calibrating":
		m.State.WithLabelValues(device_mac, coverID).Set(2)
	default:
		m.State.WithLabelValues(device_mac, coverID).Set(-1)
	}

	m.APower.WithLabelValues(device_mac, coverID).Set(status.Apower)
	m.Voltage.WithLabelValues(device_mac, coverID).Set(status.Voltage)
	m.Current.WithLabelValues(device_mac, coverID).Set(status.Current)
	m.Pf.WithLabelValues(device_mac, coverID).Set(status.Pf)
	m.Freq.WithLabelValues(device_mac, coverID).Set(status.Freq)
	m.Energy.WithLabelValues(device_mac, coverID).Set(status.Aenergy.Total)
	m.Temperature.WithLabelValues(device_mac, coverID, "dC").Set(status.Temperature.TC)
	m.Temperature.WithLabelValues(device_mac, coverID, "dF").Set(status.Temperature.TF)
	m.PosControl.WithLabelValues(device_mac, coverID).Set(boolToFloat64(status.PosControl))
	m.Position.WithLabelValues(device_mac, coverID).Set(float64(status.CurrentPos))
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
