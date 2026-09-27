# Shelly Prometheus Exporter

Shelly Prometheus Exporter is a Go-based application that collects metrics from Shelly devices via their REST API endpoints and exposes them in Prometheus-compatible format. This tool enables detailed monitoring and visualization of your Shelly devices in Prometheus and Grafana.

## Usage

Shelly Prometheus Exporter can be deployed in three different ways:

1. **Standalone Binary**
   - Download the latest standalone binary from the [GitHub Releases](https://github.com/supporterino/shelly_exporter/releases).
   - Ensure you have a `config.yaml` file configured with the necessary options. You can refer to the example provided in the repository.
   - Run the binary:
     ```bash
     ./shelly_exporter -config config.yaml
     ```

2. **Docker Image**
   - Use the Docker image available at `ghcr.io/supporterino/shelly_exporter`.
   - Ensure you have a `config.yaml` file configured with the necessary options.
   - Run the Docker container:
     ```bash
     docker run -v /path/to/config.yaml:/config.yaml -p 8080:8080 ghcr.io/supporterino/shelly_exporter
     ```

3. **Helm Chart**
   - Deploy using the Helm chart available at [https://supporterino.github.io/shelly_exporter](https://supporterino.github.io/shelly_exporter) with the chart name `shelly-exporter`.
   - Customize the configuration through the chart's `values.yaml`.
   - Install the chart:
     ```bash
     helm repo add supporterino https://supporterino.github.io/shelly_exporter
     helm install my-shelly-exporter supporterino/shelly-exporter -f values.yaml
     ```

### Configuration

- **Standalone Binary and Docker Image**: The exporter reads a YAML configuration file. The path is supplied with the `-config` flag and defaults to `config.yaml`. Unknown fields are rejected at startup so typos fail fast.
- **Helm Chart**: Configuration is managed through the `values.yaml` file, allowing fine-tuned customization of the deployment.

```yaml
listenAddress: :8080
debug: false
deviceUpdateInterval: 30s
devices:
  - host: 10.1.255.111
  - host: 10.1.255.112
    username: admin
    password: secret
```

- `listenAddress` — HTTP listen address (default `:8080`).
- `debug` — emit debug-level logs (default `false`).
- `deviceUpdateInterval` — polling period as a Go duration string such as `30s` or `1m` (default `30s`). A bare integer is rejected.
- `devices` — list of devices to poll, each with a `host` and optional `username`/`password`. When credentials are configured, the exporter authenticates using the scheme the device advertises (`WWW-Authenticate`): HTTP Basic or HTTP Digest (MD5 or SHA-256).

## Metrics

The exporter fetches metrics from various RPC calls in the Shelly API. Below are the exposed metric families with their exact Prometheus names and labels. Contributions for additional metrics are welcome.

### Supported models

Model-specific metrics are collected by the `app` value reported by `Shelly.GetDeviceInfo`:

| `app`       | Collected metric families                                   |
|-------------|-------------------------------------------------------------|
| `Plus2PM`   | Cover metrics when `profile` is `cover`, otherwise switch metrics (covers both channels). |
| `PlusPlugS` | Switch metrics.                                             |
| `Mini1G3`   | Switch metrics.                                             |
| `Pro4PM`    | Switch metrics for every channel (`switch_id` = `0`–`3`).   |

Devices with any other `app` value are still polled for the generic `Shelly.GetDeviceInfo`, `Shelly.GetStatus`, and `Shelly.GetConfig` families; the exporter logs that no model-specific handler is registered. Switch and cover component IDs are discovered from `Shelly.GetConfig`, so multi-channel devices are collected without extra configuration.

> Metrics for fields the device does not report are omitted (not exported as `0`). For example a relay without power metering exposes only `shelly_switch_state` and `shelly_switch_temperature`.

> Switch and cover metrics carry a `name` label with the channel name configured on the device (for example `Gateway` or `NAS 1`); it is empty when the channel is unnamed.

### Shelly.GetDeviceInfo

| Metric Name             | Labels                                                                  | Explanation                                                     |
|-------------------------|-------------------------------------------------------------------------|-----------------------------------------------------------------|
| `shelly_device_info`    | `device_name`, `device_id`, `device_mac`, `model`, `fw_version`, `app` | Static device information exposed as labels.                    |
| `shelly_device_auth`    | `device_mac`                                                            | Indicates if authentication is enabled (1 for true, 0 for false). |

### Shelly.GetStatus

| Metric Name             | Labels                       | Explanation                              |
|-------------------------|------------------------------|------------------------------------------|
| `shelly_system_uptime`  | `device_mac`                 | System uptime in seconds.                |
| `shelly_system_ram`     | `device_mac`, `kind`         | RAM sizes free and used in bytes.        |
| `shelly_system_fs`      | `device_mac`, `kind`         | FS sizes free and used in bytes.         |
| `shelly_system_wifi_rssi` | `device_mac`, `ssid`, `sta_ip` | Wi-Fi RSSI signal strength in dBm.     |

### Shelly.GetConfig

| Metric Name                              | Labels               | Explanation                                                       |
|------------------------------------------|----------------------|-------------------------------------------------------------------|
| `shelly_device_ble`                      | `device_mac`         | Indicates if BLE is enabled (1 for true, 0 for false).            |
| `shelly_device_cloud`                    | `device_mac`         | Indicates if Cloud is enabled (1 for true, 0 for false).          |
| `shelly_device_cloud_server`             | `device_mac`, `server` | Cloud server configuration (labels include the server address).  |
| `shelly_device_eth`                      | `device_mac`         | Indicates if Ethernet is enabled (1 for true, 0 for false).       |
| `shelly_device_eth_ipv4_mode`            | `device_mac`, `mode` | Ethernet IPv4 mode (for example `dhcp`). Only exported when the device reports an Ethernet IPv4 mode. |
| `shelly_device_wifi_ap`                  | `device_mac`         | Indicates if Wi-Fi AP is enabled (1 for true, 0 for false).       |
| `shelly_device_wifi_sta`                 | `device_mac`         | Indicates if Wi-Fi STA is enabled (1 for true, 0 for false).      |
| `shelly_device_wifi_roaming_rssi_threshold` | `device_mac`      | RSSI threshold for Wi-Fi roaming.                                 |

### Cover.GetStatus

| Metric Name               | Labels                                                | Explanation                                                                                              |
|---------------------------|-------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| `shelly_cover_state`      | `device_mac`, `cover_id`, `name`                      | Current cover state (1 = open, 0 = closed, 2 = in movement, 3 = stopped, -1 = unknown).                  |
| `shelly_cover_power`      | `device_mac`, `cover_id`, `name`                      | Active power in Watts.                                                                                   |
| `shelly_cover_voltage`    | `device_mac`, `cover_id`, `name`                      | Present voltage in Volts.                                                                                |
| `shelly_cover_current`    | `device_mac`, `cover_id`, `name`                      | Current draw in Amps.                                                                                    |
| `shelly_cover_powerfactor`| `device_mac`, `cover_id`, `name`                      | Power factor.                                                                                            |
| `shelly_cover_frequency`  | `device_mac`, `cover_id`, `name`                      | Input frequency in Hz.                                                                                   |
| `shelly_cover_energy`     | `device_mac`, `cover_id`, `name`                      | Total consumption in Wh.                                                                                 |
| `shelly_cover_temperature`| `device_mac`, `cover_id`, `name`, `temperature_unit`  | Temperature (`temperature_unit` is `dC` or `dF`).                                                        |
| `shelly_cover_pos_control`| `device_mac`, `cover_id`, `name`                      | Whether position control is present.                                                                     |
| `shelly_cover_position`   | `device_mac`, `cover_id`, `name`                      | Current position of the cover.                                                                           |

### Switch.GetStatus

| Metric Name                | Labels                                                 | Explanation                       |
|----------------------------|--------------------------------------------------------|-----------------------------------|
| `shelly_switch_state`      | `device_mac`, `switch_id`, `name`                      | Current switch state.             |
| `shelly_switch_power`      | `device_mac`, `switch_id`, `name`                      | Active power in Watts.            |
| `shelly_switch_voltage`    | `device_mac`, `switch_id`, `name`                      | Present voltage in Volts.         |
| `shelly_switch_current`    | `device_mac`, `switch_id`, `name`                      | Current draw in Amps.             |
| `shelly_switch_frequency`  | `device_mac`, `switch_id`, `name`                      | Input frequency in Hz.            |
| `shelly_switch_energy`     | `device_mac`, `switch_id`, `name`                      | Total consumption in Wh.          |
| `shelly_switch_temperature`| `device_mac`, `switch_id`, `name`, `temperature_unit`  | Temperature (`dC` or `dF`).       |

> The metering families (`power`, `voltage`, `current`, `frequency`, `energy`) are only exported for switches that report them; relays without power measurement omit them.

### Switch.GetConfig

| Metric Name                          | Labels                                 | Explanation                                  |
|--------------------------------------|----------------------------------------|----------------------------------------------|
| `shelly_switch_initial_state`        | `device_mac`, `switch_id`, `name`       | Initial state of the switch after power loss (`1` = `on`, `0` = `off`, `2` = `restore_last`, `3` = `match_input`, `-1` = unknown). |
| `shelly_switch_auto_on`              | `device_mac`, `switch_id`, `name`, `delay` | Auto on behavior of the switch.        |
| `shelly_switch_auto_off`             | `device_mac`, `switch_id`, `name`, `delay` | Auto off behavior of the switch.       |
| `shelly_switch_recover_volate_errors`| `device_mac`, `switch_id`, `name`       | Behavior after voltage errors.               |
| `shelly_switch_power_limit`          | `device_mac`, `switch_id`, `name`       | Power limit in Watts.                        |
| `shelly_switch_voltage_limit`        | `device_mac`, `switch_id`, `name`, `kind` | Voltage limits (`kind` is `overvoltage` or `undervoltage`). |
| `shelly_switch_current_limit`        | `device_mac`, `switch_id`, `name`       | Current limit in Amps.                       |

> Configuration families are only exported for settings the device reports; for example a device that has no `current_limit` does not expose `shelly_switch_current_limit`.

### WiFi.GetStatus

| Metric Name          | Labels                      | Explanation                       |
|----------------------|-----------------------------|-----------------------------------|
| `shelly_wifi_status` | `device_mac`, `status`, `ip`| Status of the Wi-Fi connection.   |
| `shelly_wifi_ssid`   | `device_mac`, `ssid`        | SSID of the Wi-Fi network.        |
| `shelly_wifi_rssi`   | `device_mac`               | Wi-Fi RSSI signal strength in dBm.|

### Scrape success

| Metric Name  | Labels                 | Explanation                                                                                                          |
|--------------|------------------------|----------------------------------------------------------------------------------------------------------------------|
| `shelly_up`  | `device_mac`, `host`   | `1` when the device's most recent poll cycle succeeded, `0` otherwise. The `host` label is always present; `device_mac` is empty until an offline-at-startup device reports its MAC. |

## Contributing

We welcome contributions to this project! Feel free to:

* Open issues for bug reports or feature requests.
* Submit pull requests with enhancements or fixes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
