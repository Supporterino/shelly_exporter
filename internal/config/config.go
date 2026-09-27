package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath is the path used when the -config flag is omitted.
const DefaultConfigPath = "config.yaml"

// Config is the decoded exporter configuration.
type Config struct {
	ListenAddress        string   `yaml:"listenAddress"`
	Debug                bool     `yaml:"debug"`
	DeviceUpdateInterval Duration `yaml:"deviceUpdateInterval"`
	Devices              []Device `yaml:"devices"`
}

// Device is a configured Shelly device.
type Device struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Duration is a YAML type that accepts a Go duration string such as "30s" or
// "1m" and rejects bare numbers.
type Duration time.Duration

// UnmarshalYAML parses a Go duration string.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
		return fmt.Errorf("invalid duration %q: a duration string such as \"30s\" is required", value.Value)
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	*d = Duration(parsed)
	return nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Default returns a configuration populated with the documented defaults.
func Default() Config {
	return Config{
		ListenAddress:        ":8080",
		Debug:                false,
		DeviceUpdateInterval: Duration(30 * time.Second),
	}
}

// Load reads, decodes, and validates the configuration at path.
func Load(path string) (*Config, error) {
	if err := ValidateConfigPath(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	cfg := Default()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate reports whether the configuration is usable.
func (c *Config) Validate() error {
	if len(c.Devices) == 0 {
		return fmt.Errorf("invalid config: at least one device is required")
	}

	if c.DeviceUpdateInterval.Duration() <= 0 {
		return fmt.Errorf("invalid config: deviceUpdateInterval must be positive")
	}

	seen := make(map[string]struct{}, len(c.Devices))
	for i := range c.Devices {
		host := strings.TrimSpace(c.Devices[i].Host)
		if host == "" {
			return fmt.Errorf("invalid config: device %d has an empty host", i)
		}
		if _, exists := seen[host]; exists {
			return fmt.Errorf("invalid config: device %d duplicates host %q", i, host)
		}
		seen[host] = struct{}{}
		c.Devices[i].Host = host
	}

	return nil
}

// ParseFlags parses the command line flags and returns the configuration path.
func ParseFlags() (string, error) {
	configPath, err := parseFlags(os.Args[1:])
	if err != nil {
		return "", err
	}

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func parseFlags(args []string) (string, error) {
	fs := flag.NewFlagSet("shelly_exporter", flag.ContinueOnError)
	configPath := fs.String("config", DefaultConfigPath, "path to config file")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	return *configPath, nil
}

// ValidateConfigPath makes sure the provided path is a readable file.
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
