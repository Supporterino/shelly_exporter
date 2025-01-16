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

- **Standalone Binary and Docker Image**: Requires a `config.yaml` file that defines the necessary settings. Refer to the example in the repository for configuration options.
- **Helm Chart**: Configuration is managed through the `values.yaml` file, allowing fine-tuned customization of the deployment.

## Metrics

The exporter fetches metrics from various RPC calls in the Shelly API. Below are the exposed metrics. Contributions for additional metrics are welcome.

### Shelly.GetConfig

| Metric Name                     | Labels                              | Example                                                                 | Explanation                                                            |
|---------------------------------|-------------------------------------|-------------------------------------------------------------------------|------------------------------------------------------------------------|
| `ble_enabled`                   | `device_mac`                       | `ble_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 1`                        | Indicates if BLE is enabled (1 for true, 0 for false).                |
| `cloud_enabled`                 | `device_mac`                       | `cloud_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 0`                      | Indicates if Cloud is enabled (1 for true, 0 for false).              |
| `cloud_server_info`             | `device_mac`, `server`             | `cloud_server_info{device_mac="AA:BB:CC:DD:EE:FF", server="example.com"} 1` | Provides cloud server configuration (e.g., server address).           |
| `eth_enabled`                   | `device_mac`                       | `eth_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 1`                        | Indicates if Ethernet is enabled (1 for true, 0 for false).           |
| `eth_ipv4_mode`                 | `device_mac`, `mode`               | `eth_ipv4_mode{device_mac="AA:BB:CC:DD:EE:FF", mode="dhcp"} 1`         | Reports the IPv4 mode of Ethernet (e.g., `dhcp`, `static`).           |
| `input_inverted`                | `device_mac`, `input_id`, `type`   | `input_inverted{device_mac="AA:BB:CC:DD:EE:FF", input_id="1", type="digital"} 1` | Shows the state of inputs, including type and ID.                     |
| `switch_auto_on`                | `device_mac`, `switch_id`          | `switch_auto_on{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 1`      | Indicates if the automatic on feature is enabled for a switch.        |
| `switch_auto_on_delay`          | `device_mac`, `switch_id`          | `switch_auto_on_delay{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 30` | Reports the delay (in seconds) before the switch auto-on feature activates. |
| `switch_auto_off_delay`         | `device_mac`, `switch_id`          | `switch_auto_off_delay{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 60` | Reports the delay (in seconds) before the switch auto-off feature activates. |
| `switch_power_limit`            | `device_mac`, `switch_id`          | `switch_power_limit{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 100` | Specifies the power limit (in watts) for a switch.                    |
| `wifi_ap_enabled`               | `device_mac`                       | `wifi_ap_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 1`                    | Indicates if the Wi-Fi Access Point (AP) is enabled (1 for true, 0 for false). |
| `wifi_sta_enabled`              | `device_mac`                       | `wifi_sta_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 0`                   | Indicates if the Wi-Fi Station (STA) mode is enabled (1 for true, 0 for false). |
| `wifi_roaming_rssi_threshold`   | `device_mac`                       | `wifi_roaming_rssi_threshold{device_mac="AA:BB:CC:DD:EE:FF"} -75`      | Reports the RSSI threshold for triggering Wi-Fi roaming.              |

### Shelly.GetStatus

| Metric Name             | Labels                                    | Example                                                                                       | Explanation                                                           |
|-------------------------|-------------------------------------------|-----------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| `input_state`           | `device_mac`, `input_id`                 | `input_state{device_mac="AA:BB:CC:DD:EE:FF", input_id="1"} 1`                                 | Indicates the state of a specific input (e.g., on/off).               |
| `switch_state`          | `device_mac`, `switch_id`                | `switch_state{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 1`                               | Indicates the state of a switch (e.g., on/off).                       |
| `switch_apower`         | `device_mac`, `switch_id`                | `switch_apower{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 150`                            | Apparent power of the switch in watts.                                |
| `switch_voltage`        | `device_mac`, `switch_id`                | `switch_voltage{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 230`                           | Voltage level of the switch in volts.                                 |
| `switch_current`        | `device_mac`, `switch_id`                | `switch_current{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 0.65`                          | Current drawn by the switch in amperes.                               |
| `switch_energy`         | `device_mac`, `switch_id`                | `switch_energy{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 12.5`                           | Energy consumption of the switch in kilowatt-hours.                   |
| `switch_temperature`    | `device_mac`, `switch_id`                | `switch_temperature{device_mac="AA:BB:CC:DD:EE:FF", switch_id="1"} 45`                        | Temperature of the switch in degrees Celsius.                         |
| `system_uptime`         | `device_mac`                             | `system_uptime{device_mac="AA:BB:CC:DD:EE:FF"} 3600`                                          | System uptime in seconds.                                             |
| `system_ram_free`       | `device_mac`                             | `system_ram_free{device_mac="AA:BB:CC:DD:EE:FF"} 1048576`                                     | Amount of free RAM in bytes.                                          |
| `system_ram_size`       | `device_mac`                             | `system_ram_size{device_mac="AA:BB:CC:DD:EE:FF"} 2097152`                                     | Total RAM size in bytes.                                              |
| `system_fs_free`        | `device_mac`                             | `system_fs_free{device_mac="AA:BB:CC:DD:EE:FF"} 524288`                                       | Amount of free filesystem space in bytes.                             |
| `system_fs_size`        | `device_mac`                             | `system_fs_size{device_mac="AA:BB:CC:DD:EE:FF"} 1048576`                                      | Total filesystem size in bytes.                                       |
| `wifi_rssi`             | `device_mac`, `ssid`, `sta_ip`           | `wifi_rssi{device_mac="AA:BB:CC:DD:EE:FF", ssid="MySSID", sta_ip="192.168.1.2"} -65`          | Wi-Fi RSSI signal strength in dBm.                                    |
| `update_available`      | `device_mac`, `version`                  | `update_available{device_mac="AA:BB:CC:DD:EE:FF", version="1.0.0"} 1`                         | Indicates if a firmware update is available (1 for yes, 0 for no).    |

### Shelly.GetDeviceInfo

| Metric Name      | Labels                                                      | Example                                                                                                                   | Explanation                                                      |
|------------------|-------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------|
| `device_info`    | `device_name`, `device_id`, `device_mac`, `model`, `fw_version`, `app` | `device_info{device_name="Device1", device_id="12345", device_mac="AA:BB:CC:DD:EE:FF", model="Shelly1", fw_version="1.2.3", app="shelly"} 1` | Exposes static device information as labels such as model, firmware version, and application. |
| `auth_enabled`   | `device_mac`                                                | `auth_enabled{device_mac="AA:BB:CC:DD:EE:FF"} 1`                                                                           | Indicates whether authentication is enabled on the device (1 for true, 0 for false).           |

## Contributing

We welcome contributions to this project! Feel free to:

* Open issues for bug reports or feature requests.
* Submit pull requests with enhancements or fixes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
