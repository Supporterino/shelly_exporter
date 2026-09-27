package exporter

import "github.com/supporterino/shelly_exporter/internal/client"

// Device is a device to register with the DeviceManager.
type Device struct {
	Host     string
	Username string
	Password string
}

// deviceState is the per-device polling state. It is owned and mutated only by
// the device's own polling goroutine.
type deviceState struct {
	host    string
	fetcher client.Fetcher

	mac     string
	model   string
	profile string
}
