package metrics

import (
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/devices"
	"github.com/supporterino/shelly_exporter/devices/ShellySmartPlugS"
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register(cfg *config.YamlConfig) {
	for _, device := range cfg.Devices {
		ShellySmartPlugS.RegisterSmartPlugS(&devices.DeviceConfig{
			Host:     device.Host,
			Username: device.Username,
			Password: device.Password,
		}, cfg.DeviceUpdateInterval)
	}
}
