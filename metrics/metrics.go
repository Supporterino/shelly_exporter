package metrics

import (
	"github.com/supporterino/shelly_exporter/config"
	"github.com/supporterino/shelly_exporter/rpc"
)

// Register initializes Prometheus metrics and starts periodic API fetching.
func Register(cfg *config.YamlConfig) {
	for _, device := range cfg.Devices {
		rpc.RegisterDevice(&rpc.DeviceConfig{
			Host:     device.Host,
			Username: device.Username,
			Password: device.Password,
		}, cfg.DeviceUpdateInterval)
	}
}
