package ShellyGetDeviceInfo

import (
	"testing"

	"github.com/LukeEvansTech/shelly-prometheus-exporter/client"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// After a firmware update the plug must have ONE shelly_device_info series,
// on the new version. Two series per MAC break every dashboard join on it.
func TestFirmwareChangeLeavesOneSeries(t *testing.T) {
	RegisterShellyGetDeviceInfoMetrics()
	base := client.ShellyGetDeviceInfoResponse{Mac: "AABBCC", ID: "plug-aabbcc", App: "PlusPlugUK", Model: "SNPL-00112UK"}

	old, updated, other := base, base, base
	old.FwID = "20260710-101152/2.0.0-g87fbfa4"
	updated.FwID = "20260923-075555/2.0.1-ge1a198b"
	other.Mac, other.ID, other.FwID = "DDEEFF", "plug-ddeeff", old.FwID

	metrics.UpdateMetrics(old)
	metrics.UpdateMetrics(other)
	metrics.UpdateMetrics(updated)

	if n := testutil.CollectAndCount(metrics.DeviceInfo); n != 2 {
		t.Fatalf("shelly_device_info series = %d, want 2 (one per plug)", n)
	}
	if v := testutil.ToFloat64(metrics.DeviceInfo.WithLabelValues("", base.ID, base.Mac, base.Model, updated.FwID, base.App)); v != 1 {
		t.Errorf("new-firmware series = %v, want 1", v)
	}
}
