package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// UpCollector exposes the per-device scrape-success gauge.
type UpCollector struct {
	up *prometheus.GaugeVec
}

// NewUpCollector constructs and registers an UpCollector.
func NewUpCollector(reg prometheus.Registerer) (*UpCollector, error) {
	c := &UpCollector{
		up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Name:      "up",
			Help:      "Whether the device's most recent poll cycle succeeded (1) or failed (0).",
		}, []string{"device_mac", "host"}),
	}

	if err := register(reg, c.up); err != nil {
		return nil, fmt.Errorf("failed to register up collector: %w", err)
	}

	return c, nil
}

// Set records the scrape success for a device.
func (c *UpCollector) Set(value float64, deviceMAC, host string) {
	c.up.WithLabelValues(deviceMAC, host).Set(value)
}

// Delete removes a previously recorded series.
func (c *UpCollector) Delete(deviceMAC, host string) bool {
	return c.up.DeleteLabelValues(deviceMAC, host)
}
