package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

const validConfig = `listenAddress: :9090
deviceUpdateInterval: 15s
devices:
  - host: 10.0.0.1
`

func TestLoadValidConfig(t *testing.T) {
	cfg, err := Load(writeConfig(t, validConfig))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.ListenAddress != ":9090" {
		t.Errorf("ListenAddress = %q, want :9090", cfg.ListenAddress)
	}
	if cfg.DeviceUpdateInterval.Duration() != 15*time.Second {
		t.Errorf("DeviceUpdateInterval = %s, want 15s", cfg.DeviceUpdateInterval.Duration())
	}
	if len(cfg.Devices) != 1 || cfg.Devices[0].Host != "10.0.0.1" {
		t.Errorf("Devices = %+v, want a single 10.0.0.1", cfg.Devices)
	}
}

func TestLoadTrimsHostWhitespace(t *testing.T) {
	cfg, err := Load(writeConfig(t, "deviceUpdateInterval: 15s\ndevices:\n  - host: \" 10.0.0.1 \"\n"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got := cfg.Devices[0].Host; got != "10.0.0.1" {
		t.Errorf("Host = %q, want trimmed 10.0.0.1", got)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	_, err := Load(writeConfig(t, `deviceUpdateInterval: 15s
unknownField: nope
devices:
  - host: 10.0.0.1
`))
	if err == nil {
		t.Fatal("expected unknown field to be rejected")
	}
	if !strings.Contains(err.Error(), "unknownField") {
		t.Errorf("error %q does not mention the offending field", err)
	}
}

func TestDurationFormat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "seconds", value: "30s", want: 30 * time.Second},
		{name: "minutes", value: "1m", want: time.Minute},
		{name: "bare integer", value: "30", wantErr: true},
		{name: "garbage", value: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(writeConfig(t, "deviceUpdateInterval: "+tt.value+"\ndevices:\n  - host: 10.0.0.1\n"))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.DeviceUpdateInterval.Duration() != tt.want {
				t.Errorf("interval = %s, want %s", cfg.DeviceUpdateInterval.Duration(), tt.want)
			}
		})
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := Load(writeConfig(t, "devices:\n  - host: 10.0.0.1\n"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.ListenAddress != ":8080" {
		t.Errorf("ListenAddress = %q, want :8080", cfg.ListenAddress)
	}
	if cfg.DeviceUpdateInterval.Duration() != 30*time.Second {
		t.Errorf("DeviceUpdateInterval = %s, want 30s", cfg.DeviceUpdateInterval.Duration())
	}
	if cfg.Debug {
		t.Error("Debug = true, want false by default")
	}
}

func TestDebugFlagLoaded(t *testing.T) {
	cfg, err := Load(writeConfig(t, "debug: true\ndevices:\n  - host: 10.0.0.1\n"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
}

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "no devices", content: "deviceUpdateInterval: 30s\ndevices: []\n"},
		{name: "empty host", content: "deviceUpdateInterval: 30s\ndevices:\n  - host: \"\"\n"},
		{name: "zero interval", content: "deviceUpdateInterval: 0s\ndevices:\n  - host: 10.0.0.1\n"},
		{name: "negative interval", content: "deviceUpdateInterval: -1s\ndevices:\n  - host: 10.0.0.1\n"},
		{name: "duplicate host", content: "deviceUpdateInterval: 30s\ndevices:\n  - host: 10.0.0.1\n  - host: 10.0.0.1\n"},
		{name: "duplicate host with whitespace", content: "deviceUpdateInterval: 30s\ndevices:\n  - host: 10.0.0.1\n  - host: \" 10.0.0.1 \"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Load(writeConfig(t, tt.content)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected error for a missing file")
	}
}

func TestParseFlagsDefaultAndOverride(t *testing.T) {
	path, err := parseFlags(nil)
	if err != nil {
		t.Fatalf("parseFlags returned error: %v", err)
	}
	if path != DefaultConfigPath {
		t.Errorf("default path = %q, want %q", path, DefaultConfigPath)
	}

	path, err = parseFlags([]string{"-config", "other.yaml"})
	if err != nil {
		t.Fatalf("parseFlags returned error: %v", err)
	}
	if path != "other.yaml" {
		t.Errorf("path = %q, want other.yaml", path)
	}
}
