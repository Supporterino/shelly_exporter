// Package labelset holds gauges whose labels describe a device's CURRENT
// state (firmware version, SSID, IP, cloud server), where each device must
// have exactly one series at a time.
//
// A plain GaugeVec keeps every label set it has ever seen, so when a
// device's firmware or IP changes the old series stays exported forever,
// reporting 1 for a state the device left long ago. Queries that join on
// device_mac then hit two series per device and fail.
package labelset

import (
	"slices"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Gauge is a GaugeVec that keeps one series per key (the device MAC). It is
// a prometheus.Collector, so it registers like the GaugeVec it wraps.
type Gauge struct {
	*prometheus.GaugeVec

	mu   sync.Mutex
	last map[string][]string
}

// New wraps vec.
func New(vec *prometheus.GaugeVec) *Gauge {
	return &Gauge{GaugeVec: vec, last: map[string][]string{}}
}

// Set sets the series for labelValues and deletes key's previous series if
// its label values differ.
func (g *Gauge) Set(key string, value float64, labelValues ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if prev, ok := g.last[key]; ok && !slices.Equal(prev, labelValues) {
		g.DeleteLabelValues(prev...)
	}
	g.WithLabelValues(labelValues...).Set(value)
	g.last[key] = slices.Clone(labelValues)
}
