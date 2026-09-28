package ShellyGetDeviceInfo

import (
	"fmt"

	"github.com/LukeEvansTech/shelly-prometheus-exporter/client"
	"github.com/LukeEvansTech/shelly-prometheus-exporter/labelset"
	"github.com/prometheus/client_golang/prometheus"
)

type ShellyGetDeviceInfoMetrics struct {
	DeviceInfo  *labelset.Gauge
	AuthEnabled *prometheus.GaugeVec
}

var metrics *ShellyGetDeviceInfoMetrics

func RegisterShellyGetDeviceInfoMetrics() {
	metrics = &ShellyGetDeviceInfoMetrics{
		DeviceInfo: labelset.New(prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "info",
			Help:      "Static device information exposed as labels (model, firmware version, app).",
		}, []string{"device_name", "device_id", "device_mac", "model", "fw_version", "app"})),
		AuthEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shelly",
			Subsystem: "device",
			Name:      "auth",
			Help:      "Indicates if authentication is enabled on the device.",
		}, []string{"device_mac"}),
	}

	prometheus.MustRegister(
		metrics.AuthEnabled,
		metrics.DeviceInfo,
	)
}

// UpdateShellyGetDeviceInfoMetrics fetches device info, updates its metrics
// and returns it. The caller uses the returned value for this device's
// identity: every device's goroutine shares this package, so state kept here
// would belong to whichever device answered last.
func UpdateShellyGetDeviceInfoMetrics(apiClient *client.APIClient) (client.ShellyGetDeviceInfoResponse, error) {
	var info client.ShellyGetDeviceInfoResponse
	err := apiClient.FetchData("/rpc/Shelly.GetDeviceInfo", &info)
	if err != nil {
		return info, fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(info)

	return info, nil
}

func (m *ShellyGetDeviceInfoMetrics) UpdateMetrics(info client.ShellyGetDeviceInfoResponse) {
	m.DeviceInfo.Set(info.Mac, 1, info.Name, info.ID, info.Mac, info.Model, info.FwID, info.App)
	m.AuthEnabled.WithLabelValues(info.Mac).Set(boolToFloat64(info.AuthEn))
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
