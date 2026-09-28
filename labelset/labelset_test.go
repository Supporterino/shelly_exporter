package labelset

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newGauge() *Gauge {
	return New(prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "t"}, []string{"device_mac", "fw_version"}))
}

// A firmware change must replace the device's series, not add a second one.
func TestChangedLabelsReplaceTheSeries(t *testing.T) {
	g := newGauge()
	g.Set("aa", 1, "aa", "2.0.0")
	g.Set("aa", 1, "aa", "2.0.1")
	if n := testutil.CollectAndCount(g); n != 1 {
		t.Fatalf("series = %d, want 1 after a label change", n)
	}
	if v := testutil.ToFloat64(g.WithLabelValues("aa", "2.0.1")); v != 1 {
		t.Errorf("current series value = %v, want 1", v)
	}
}

// Other devices are untouched, and repeating the same labels is a no-op.
func TestDevicesAreIndependent(t *testing.T) {
	g := newGauge()
	g.Set("aa", 1, "aa", "2.0.0")
	g.Set("bb", 1, "bb", "2.0.0")
	g.Set("aa", 1, "aa", "2.0.0")
	g.Set("aa", 1, "aa", "2.0.1")
	if n := testutil.CollectAndCount(g); n != 2 {
		t.Fatalf("series = %d, want 2 (one per device)", n)
	}
}

func TestConcurrentSet(t *testing.T) {
	g := newGauge()
	done := make(chan struct{})
	for i := range 8 {
		go func() {
			for j := range 200 {
				g.Set("aa", 1, "aa", []string{"a", "b"}[(i+j)%2])
			}
			done <- struct{}{}
		}()
	}
	for range 8 {
		<-done
	}
	if n := testutil.CollectAndCount(g); n != 1 {
		t.Fatalf("series = %d, want 1 under concurrent updates", n)
	}
}
