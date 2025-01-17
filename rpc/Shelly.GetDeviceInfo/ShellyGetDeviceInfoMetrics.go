package ShellyGetDeviceInfo

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type GetDeviceInfoMetrics struct {
	DeviceInfo  *prometheus.GaugeVec
	AuthEnabled *prometheus.GaugeVec
}

var metrics *GetDeviceInfoMetrics

func RegisterShelly_GetDeviceInfoMetrics() {
	metrics = &GetDeviceInfoMetrics{
		DeviceInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "device_info",
			Help: "Static device information exposed as labels (model, firmware version, app).",
		}, []string{"device_name", "device_id", "device_mac", "model", "fw_version", "app"}),
		AuthEnabled: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "auth_enabled",
			Help: "Indicates if authentication is enabled on the device.",
		}, []string{"device_mac"}),
	}

	prometheus.MustRegister(
		metrics.AuthEnabled,
		metrics.DeviceInfo,
	)
}

func UpdateShelly_GetDeviceInfoMetrics(apiClient *client.APIClient) error {
	var info client.ShellyGetDeviceInfoResponse
	err := apiClient.FetchData("/rpc/Shelly.GetDeviceInfo", &info)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(info)

	return nil
}

func (m *GetDeviceInfoMetrics) UpdateMetrics(info client.ShellyGetDeviceInfoResponse) {
	m.DeviceInfo.WithLabelValues(info.Name, info.ID, info.MAC, info.Model, info.FwID, info.App).Set(1)
	m.AuthEnabled.WithLabelValues(info.MAC).Set(boolToFloat64(info.AuthEn))
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
