package metrics

import (
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/rpc"
	CoverGetStatus "github.com/supporterino/shelly_exporter/rpc/Cover.GetStatus"
	ShellyGetConfig "github.com/supporterino/shelly_exporter/rpc/Shelly.GetConfig"
	ShellyGetDeviceInfo "github.com/supporterino/shelly_exporter/rpc/Shelly.GetDeviceInfo"
	ShellyGetStatus "github.com/supporterino/shelly_exporter/rpc/Shelly.GetStatus"
	SwitchGetConfig "github.com/supporterino/shelly_exporter/rpc/Switch.GetConfig"
	SwitchGetStatus "github.com/supporterino/shelly_exporter/rpc/Switch.GetStatus"
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register(cfg *config.YamlConfig, cfgPath *string) {
	ShellyGetConfig.RegisterShellyGetConfigMetrics()
	ShellyGetStatus.RegisterShellyGetStatusMetrics()
	ShellyGetDeviceInfo.RegisterShellyGetDeviceInfoMetrics()
	CoverGetStatus.RegisterCoverGetStatusMetrics()
	SwitchGetStatus.RegisterSwitchGetStatusMetrics()
	SwitchGetConfig.RegisterSwitchGetConfigMetrics()

	dm := rpc.NewDeviceManager()

	for _, device := range cfg.Devices {
		dm.RegisterDevice(&rpc.DeviceConfig{
			Host:     device.Host,
			Username: device.Username,
			Password: device.Password,
		}, cfg.DeviceUpdateInterval)
	}
}
