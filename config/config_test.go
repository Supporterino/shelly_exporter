package config

import "testing"

func TestApplyCredentialDefaultsEnvFallback(t *testing.T) {
	t.Setenv("SHELLY_USERNAME", "")
	t.Setenv("SHELLY_PASSWORD", "envpass")

	cfg := &YamlConfig{Devices: []DeviceYamlConfig{
		{Host: "a"}, // empty -> env pass, admin user
		{Host: "b", Username: "u", Password: "p"}, // explicit -> preserved
		{Host: "c", Password: "explicit"},         // pass kept, user -> admin
		{Host: "d", Username: "custom"},           // explicit user, no pass -> env pass, user kept
	}}
	applyCredentialDefaults(cfg)

	if cfg.Devices[0].Password != "envpass" || cfg.Devices[0].Username != "admin" {
		t.Errorf("device a: got %+v, want password=envpass username=admin", cfg.Devices[0])
	}
	if cfg.Devices[1].Username != "u" || cfg.Devices[1].Password != "p" {
		t.Errorf("device b: explicit creds not preserved: %+v", cfg.Devices[1])
	}
	if cfg.Devices[2].Password != "explicit" || cfg.Devices[2].Username != "admin" {
		t.Errorf("device c: got %+v, want password=explicit username=admin", cfg.Devices[2])
	}
	if cfg.Devices[3].Username != "custom" || cfg.Devices[3].Password != "envpass" {
		t.Errorf("device d: got %+v, want username=custom password=envpass", cfg.Devices[3])
	}
}

func TestApplyCredentialDefaultsExplicitEnvUser(t *testing.T) {
	t.Setenv("SHELLY_USERNAME", "operator")
	t.Setenv("SHELLY_PASSWORD", "envpass")

	cfg := &YamlConfig{Devices: []DeviceYamlConfig{{Host: "a"}}}
	applyCredentialDefaults(cfg)

	if cfg.Devices[0].Username != "operator" || cfg.Devices[0].Password != "envpass" {
		t.Errorf("got %+v, want username=operator password=envpass", cfg.Devices[0])
	}
}

func TestApplyCredentialDefaultsNoEnvLeavesEmpty(t *testing.T) {
	t.Setenv("SHELLY_USERNAME", "")
	t.Setenv("SHELLY_PASSWORD", "")

	cfg := &YamlConfig{Devices: []DeviceYamlConfig{{Host: "a"}}}
	applyCredentialDefaults(cfg)

	if cfg.Devices[0].Username != "" || cfg.Devices[0].Password != "" {
		t.Errorf("expected empty creds with no env, got %+v", cfg.Devices[0])
	}
}
