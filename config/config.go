package config

type Config struct {
	ListenAddress  string `env:"LISTEN_ADDR" envDefault:":8080"`
	DeviceAddress  string `env:"DEVICE_ADDR"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	UpdateInterval int64  `env:"INTERVAL" envDefault:"30"`
}
