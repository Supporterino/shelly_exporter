# Shelly Prometheus Exporter

> Fork of [Supporterino/shelly_exporter](https://github.com/Supporterino/shelly_exporter) with dynamic component discovery.

Shelly Prometheus Exporter is a Go-based application that collects metrics from Shelly devices via their RPC API and exposes them in Prometheus-compatible format.

## Fork Changes

The upstream exporter hard-codes supported device types (`PlusPlugS`, `Plus2PM`, `Mini1G3`), meaning any other Shelly device (e.g. `PlusPlugUK`, `PlugSG3`, Gen4 devices) silently produces no switch/cover metrics.

This fork replaces the device-type switch statement with **dynamic component discovery**: on registration, the exporter parses the keys from `Shelly.GetStatus` (e.g. `switch:0`, `cover:0`) to determine which components a device has. This means **all Gen2, Gen3, and Gen4 devices work automatically** without code changes per device type.

### Changes from upstream

- **Dynamic component discovery** (`client/api_client.go`) - parses `Shelly.GetStatus` response keys to find `switch:N` and `cover:N` components
- **Generic metrics collection** (`rpc/main.go`) - iterates over discovered components instead of matching device types; WiFi metrics always collected
- **HTTP digest auth** (`client/api_client.go`) - implements Shelly Gen2 digest authentication (SHA-256, user `admin`). The `username`/`password` config fields were previously inert; they now authenticate, and a global credential can be supplied via environment variables (see Authentication).
- **Dockerfile** - removed bundled `config.yaml`, clean entrypoint accepting `--config` flag
- **CI workflows** - use `GITHUB_TOKEN` instead of custom `TOKEN` secret

## Usage

### Docker

```bash
docker run -v /path/to/config.yaml:/config/config.yaml -p 8080:8080 \
  ghcr.io/lukeevanstech/shelly_exporter --config /config/config.yaml
```

### Configuration

```yaml
listenAddress: :8080
debug: false
deviceUpdateInterval: 30
devices:
  - host: 192.168.1.100
  - host: 192.168.1.101
    password: mypassword
```

| Field | Description | Default |
|---|---|---|
| `listenAddress` | Address and port to listen on | `:8080` |
| `debug` | Enable debug logging | `false` |
| `deviceUpdateInterval` | Seconds between device polls | `30` |
| `devices[].host` | Device IP address | (required) |
| `devices[].username` | Auth username (if enabled) | (empty) |
| `devices[].password` | Auth password (if enabled) | (empty) |

### Authentication

Shelly Gen2+ devices use HTTP digest auth (SHA-256, account `admin`). When a
device has auth enabled, supply the password and the exporter authenticates
automatically.

Two ways to provide credentials (per-device config wins over the environment):

- **Per-device** in `config.yaml` via `devices[].username` / `devices[].password`.
- **Global** via environment variables, applied to every device that has no
  explicit credential:

  | Env var | Description | Default |
  |---|---|---|
  | `SHELLY_USERNAME` | Auth username for all devices | `admin` (used when a password is set) |
  | `SHELLY_PASSWORD` | Auth password for all devices | (empty) |

The global option suits a fleet that shares one admin password and a config
listing only hosts (e.g. one rendered by an operator): set `SHELLY_PASSWORD`
once and every device authenticates as `admin`.

### Endpoints

| Path | Description |
|---|---|
| `/metrics` | Prometheus metrics |
| `/health` | Health check |

## Metrics

### Switch Metrics (`switch:N` components)

| Metric | Labels | Description |
|---|---|---|
| `shelly_switch_state` | `device_mac`, `switch_id` | Switch output state (1=on, 0=off) |
| `shelly_switch_power` | `device_mac`, `switch_id` | Active power in watts |
| `shelly_switch_voltage` | `device_mac`, `switch_id` | Voltage in volts |
| `shelly_switch_current` | `device_mac`, `switch_id` | Current in amps |
| `shelly_switch_frequency` | `device_mac`, `switch_id` | Input frequency in Hz |
| `shelly_switch_energy` | `device_mac`, `switch_id` | Total energy consumption in Wh |
| `shelly_switch_temperature` | `device_mac`, `switch_id`, `temperature_unit` | Device temperature (dC/dF) |
| `shelly_switch_power_limit` | `device_mac`, `switch_id` | Configured power limit in watts |
| `shelly_switch_current_limit` | `device_mac`, `switch_id` | Configured current limit in amps |
| `shelly_switch_voltage_limit` | `device_mac`, `switch_id`, `kind` | Voltage limits (overvoltage/undervoltage) |
| `shelly_switch_initial_state` | `device_mac`, `switch_id` | Initial state on boot |
| `shelly_switch_auto_on` | `device_mac`, `switch_id`, `delay` | Auto-on enabled |
| `shelly_switch_auto_off` | `device_mac`, `switch_id`, `delay` | Auto-off enabled |

### Cover Metrics (`cover:N` components)

| Metric | Labels | Description |
|---|---|---|
| `shelly_cover_state` | `device_mac`, `cover_id` | Cover state (1=open, 0=closed, 2=moving, 3=stopped) |
| `shelly_cover_power` | `device_mac`, `cover_id` | Active power in watts |
| `shelly_cover_voltage` | `device_mac`, `cover_id` | Voltage in volts |
| `shelly_cover_current` | `device_mac`, `cover_id` | Current in amps |
| `shelly_cover_powerfactor` | `device_mac`, `cover_id` | Power factor |
| `shelly_cover_frequency` | `device_mac`, `cover_id` | Input frequency in Hz |
| `shelly_cover_energy` | `device_mac`, `cover_id` | Total energy consumption in Wh |
| `shelly_cover_temperature` | `device_mac`, `cover_id`, `temperature_unit` | Device temperature (dC/dF) |
| `shelly_cover_position` | `device_mac`, `cover_id` | Current position (0-100) |
| `shelly_cover_pos_control` | `device_mac`, `cover_id` | Position control available |

### System Metrics

| Metric | Labels | Description |
|---|---|---|
| `shelly_device_info` | `device_name`, `device_id`, `device_mac`, `model`, `fw_version`, `app` | Static device information |
| `shelly_device_auth` | `device_mac` | Authentication enabled |
| `shelly_device_ble` | `device_mac` | BLE enabled |
| `shelly_device_cloud` | `device_mac` | Cloud enabled |
| `shelly_device_eth` | `device_mac` | Ethernet enabled |
| `shelly_device_wifi_sta` | `device_mac` | WiFi station enabled |
| `shelly_device_wifi_ap` | `device_mac` | WiFi AP enabled |
| `shelly_system_uptime` | `device_mac` | System uptime in seconds |
| `shelly_system_ram` | `device_mac`, `kind` | RAM free/max in bytes |
| `shelly_system_fs` | `device_mac`, `kind` | Filesystem free/max in bytes |
| `shelly_system_wifi_rssi` | `device_mac`, `ssid`, `sta_ip` | WiFi signal strength in dBm |

## Tested Devices

| Device | App Type | Status |
|---|---|---|
| Shelly Plus Plug UK | `PlusPlugUK` | Verified |
| Shelly Plus Plug S | `PlusPlugS` | Supported (upstream) |
| Shelly Plus 2PM | `Plus2PM` | Supported (upstream) |
| Shelly 1 Mini Gen3 | `Mini1G3` | Supported (upstream) |
| All other Gen2/Gen3/Gen4 with switch or cover components | Any | Should work via dynamic discovery |

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
