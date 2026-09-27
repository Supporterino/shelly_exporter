package collector

import "github.com/prometheus/client_golang/prometheus"

func register(reg prometheus.Registerer, collectors ...prometheus.Collector) error {
	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			return err
		}
	}
	return nil
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// setGauge sets a labelled gauge when the device reported the value. Absent
// fields are left unset, so no series is created and the exporter does not
// publish fabricated zero readings.
func setGauge(g *prometheus.GaugeVec, value *float64, labels ...string) {
	if value != nil {
		g.WithLabelValues(labels...).Set(*value)
	}
}

// deleteDeviceSeries removes every series carrying the device_mac label from
// the given vectors. It is used when a device starts reporting a different MAC
// so the exporter does not keep exposing stale series under the old MAC.
func deleteDeviceSeries(deviceMAC string, vecs ...*prometheus.GaugeVec) {
	for _, v := range vecs {
		v.DeletePartialMatch(prometheus.Labels{"device_mac": deviceMAC})
	}
}
