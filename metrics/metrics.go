package metrics

import (
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/rpc"
	ShellyGetConfig "github.com/supporterino/shelly_exporter/rpc/Shelly.GetConfig"
	ShellyGetDeviceInfo "github.com/supporterino/shelly_exporter/rpc/Shelly.GetDeviceInfo"
	ShellyGetStatus "github.com/supporterino/shelly_exporter/rpc/Shelly.GetStatus"
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register(cfg *config.YamlConfig, cfgPath *string) {
	ShellyGetConfig.RegisterShelly_GetConfigMetrics()
	ShellyGetStatus.RegisterShelly_GetStatusMetrics()
	ShellyGetDeviceInfo.RegisterShelly_GetDeviceInfoMetrics()

	dm := rpc.NewDeviceManager()

	for _, device := range cfg.Devices {
		dm.RegisterDevice(&rpc.DeviceConfig{
			Host:     device.Host,
			Username: device.Username,
			Password: device.Password,
		}, cfg.DeviceUpdateInterval)
	}
}
