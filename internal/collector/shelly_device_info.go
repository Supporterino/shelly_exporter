package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
)

// DeviceInfoCollector collects static device information from Shelly.GetDeviceInfo.
type DeviceInfoCollector struct {
	DeviceInfo  *prometheus.GaugeVec
	AuthEnabled *prometheus.GaugeVec
}

// NewDeviceInfoCollector constructs and registers a DeviceInfoCollector.
func NewDeviceInfoCollector(reg prometheus.Registerer) (*DeviceInfoCollector, error) {
	c := &DeviceInfoCollector{
		DeviceInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "info",
			Help:      "Static device information exposed as labels (model, firmware version, app).",
		}, []string{"device_name", "device_id", "device_mac", "model", "fw_version", "app"}),
		AuthEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "auth",
			Help:      "Indicates if authentication is enabled on the device.",
		}, []string{"device_mac"}),
	}

	if err := register(reg, c.DeviceInfo, c.AuthEnabled); err != nil {
		return nil, fmt.Errorf("failed to register device info collector: %w", err)
	}

	return c, nil
}

// Delete removes every metric series for a device.
func (c *DeviceInfoCollector) Delete(deviceMAC string) {
	deleteDeviceSeries(deviceMAC, c.DeviceInfo, c.AuthEnabled)
}

// Update fetches Shelly.GetDeviceInfo, updates the metrics, and returns the
// decoded device information.
func (c *DeviceInfoCollector) Update(ctx context.Context, fetcher client.Fetcher) (client.ShellyGetDeviceInfoResponse, error) {
	var info client.ShellyGetDeviceInfoResponse
	if err := fetcher.FetchData(ctx, "/rpc/Shelly.GetDeviceInfo", &info); err != nil {
		return info, fmt.Errorf("failed to fetch Shelly.GetDeviceInfo: %w", err)
	}

	c.UpdateMetrics(info)
	return info, nil
}

// UpdateMetrics populates the metrics from a Shelly.GetDeviceInfo response.
func (c *DeviceInfoCollector) UpdateMetrics(info client.ShellyGetDeviceInfoResponse) {
	c.DeviceInfo.WithLabelValues(info.Name, info.ID, info.Mac, info.Model, info.FwID, info.App).Set(1)
	c.AuthEnabled.WithLabelValues(info.Mac).Set(boolToFloat64(info.AuthEn))
}
