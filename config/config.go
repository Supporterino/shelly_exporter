package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	ListenAddress        string             `yaml:"listenAddress"`
	Debug                bool               `yaml:"debug"`
	DeviceUpdateInterval time.Duration      `yaml:"deviceUpdateInterval"`
	Devices              []DeviceYamlConfig `yaml:"devices"`
}

type DeviceYamlConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*YamlConfig, error) {
	// Create config structure
	config := &YamlConfig{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	applyCredentialDefaults(config)
	return config, nil
}

// applyCredentialDefaults fills empty per-device credentials from the global
// SHELLY_USERNAME / SHELLY_PASSWORD environment variables. This lets the
// operator-rendered config carry hosts only while one shared admin password is
// supplied via the environment. A device that ends up with a password but no
// username defaults to "admin" (the Shelly Gen2 account). Explicit per-device
// values in the config file always win.
func applyCredentialDefaults(cfg *YamlConfig) {
	envUser := os.Getenv("SHELLY_USERNAME")
	envPass := os.Getenv("SHELLY_PASSWORD")
	for i := range cfg.Devices {
		d := &cfg.Devices[i]
		if d.Password == "" {
			d.Password = envPass
		}
		if d.Username == "" {
			switch {
			case envUser != "":
				d.Username = envUser
			case d.Password != "":
				d.Username = "admin"
			}
		}
	}
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
