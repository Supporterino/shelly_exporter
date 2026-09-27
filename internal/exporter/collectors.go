package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/collector"
)

// Collectors holds every collector owned by the exporter.
type Collectors struct {
	ShellyStatus *collector.ShellyStatusCollector
	ShellyConfig *collector.ShellyConfigCollector
	DeviceInfo   *collector.DeviceInfoCollector
	CoverStatus  *collector.CoverStatusCollector
	SwitchStatus *collector.SwitchStatusCollector
	SwitchConfig *collector.SwitchConfigCollector
	WiFiStatus   *collector.WiFiStatusCollector
	Up           *collector.UpCollector
}

// DeleteDevice removes every metric series belonging to a device, including its
// scrape-success series. It is used when a device reports a new MAC address.
func (c *Collectors) DeleteDevice(deviceMAC, host string) {
	c.ShellyStatus.Delete(deviceMAC)
	c.ShellyConfig.Delete(deviceMAC)
	c.DeviceInfo.Delete(deviceMAC)
	c.CoverStatus.Delete(deviceMAC)
	c.SwitchStatus.Delete(deviceMAC)
	c.SwitchConfig.Delete(deviceMAC)
	c.WiFiStatus.Delete(deviceMAC)
	c.Up.Delete(deviceMAC, host)
}

// NewCollectors constructs and registers all exporter collectors.
func NewCollectors(reg prometheus.Registerer) (*Collectors, error) {
	shellyStatus, err := collector.NewShellyStatusCollector(reg)
	if err != nil {
		return nil, err
	}
	shellyConfig, err := collector.NewShellyConfigCollector(reg)
	if err != nil {
		return nil, err
	}
	deviceInfo, err := collector.NewDeviceInfoCollector(reg)
	if err != nil {
		return nil, err
	}
	coverStatus, err := collector.NewCoverStatusCollector(reg)
	if err != nil {
		return nil, err
	}
	switchStatus, err := collector.NewSwitchStatusCollector(reg)
	if err != nil {
		return nil, err
	}
	switchConfig, err := collector.NewSwitchConfigCollector(reg)
	if err != nil {
		return nil, err
	}
	wifiStatus, err := collector.NewWiFiStatusCollector(reg)
	if err != nil {
		return nil, err
	}
	up, err := collector.NewUpCollector(reg)
	if err != nil {
		return nil, err
	}

	return &Collectors{
		ShellyStatus: shellyStatus,
		ShellyConfig: shellyConfig,
		DeviceInfo:   deviceInfo,
		CoverStatus:  coverStatus,
		SwitchStatus: switchStatus,
		SwitchConfig: switchConfig,
		WiFiStatus:   wifiStatus,
		Up:           up,
	}, nil
}
