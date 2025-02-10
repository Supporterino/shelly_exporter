package ShellyGetDeviceInfo

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/client"
)

type GetDeviceInfoMetrics struct {
	DeviceInfo    *prometheus.GaugeVec
	AuthEnabled   *prometheus.GaugeVec
	DeviceModel   *string
	DeviceProfile *string
	DeviceMac     *string
}

var metrics *GetDeviceInfoMetrics

func RegisterShellyGetDeviceInfoMetrics() {
	metrics = &GetDeviceInfoMetrics{
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

	prometheus.MustRegister(
		metrics.AuthEnabled,
		metrics.DeviceInfo,
	)
}

func UpdateShellyGetDeviceInfoMetrics(apiClient *client.APIClient) error {
	var info client.ShellyGetDeviceInfoResponse
	err := apiClient.FetchData("/rpc/Shelly.GetDeviceInfo", &info)
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	metrics.UpdateMetrics(info)

	return nil
}

func (m *GetDeviceInfoMetrics) UpdateMetrics(info client.ShellyGetDeviceInfoResponse) {
	m.DeviceInfo.WithLabelValues(info.Name, info.ID, info.Mac, info.Model, info.FwID, info.App).Set(1)
	m.AuthEnabled.WithLabelValues(info.Mac).Set(boolToFloat64(info.AuthEn))
	m.DeviceModel = &info.App
	m.DeviceProfile = &info.Profile
	m.DeviceMac = &info.Mac
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func GetDeviceType() string {
	return *metrics.DeviceModel
}

func GetDeviceProfile() string {
	return *metrics.DeviceProfile
}

func GetDeviceMac() string {
	return *metrics.DeviceMac
}
