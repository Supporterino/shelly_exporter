package config

import (
	"time"
)

type Config struct {
	ListenAddress  string `env:"LISTEN_ADDR" envDefault:":8080"`
	DeviceAddress  string `env:"DEVICE_ADDR"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	UpdateInterval int64  `env:"INTERVAL" envDefault:"30"`
}

type YamlConfig struct {
	ListenAddress string `yaml:"listenAddress"`
	LogLevel string `yaml:"logLevel"`
	DeviceUpdateInterval time.Duration `yaml:"deviceUpdateInterval"`
}