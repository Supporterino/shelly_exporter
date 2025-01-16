# Shelly Prometheus Exporter

Shelly Prometheus Exporter is a Go-based application that collects metrics from Shelly devices via their REST API endpoints and exposes them in Prometheus-compatible format. This tool enables detailed monitoring and visualization of your Shelly devices in Prometheus and Grafana.

## Usage

## Metrics

### Shelly.GetStatus

# Metrics Table

| **Metric Name**              | **Description**                                                                                 | **Labels**                              | **Example**                                                                                     |
|-------------------------------|---------------------------------------------------------------------------------------------|-----------------------------------------|-------------------------------------------------------------------------------------------------|
| `shelly_input_state`          | State of each input (`1` for active, `0` for inactive).                                       | `device_mac`, `input_id`               | `shelly_input_state{device_mac="A8032ABE54DC",input_id="input:0"} 1`                           |
| `shelly_switch_output`        | Output state of each switch (`1` for active, `0` for inactive).                              | `device_mac`, `switch_id`              | `shelly_switch_output{device_mac="A8032ABE54DC",switch_id="switch:0"} 1`                      |
| `shelly_switch_apower`        | Active power of each switch in watts.                                                        | `device_mac`, `switch_id`              | `shelly_switch_apower{device_mac="A8032ABE54DC",switch_id="switch:0"} 8.9`                    |
| `shelly_switch_voltage`       | Voltage of each switch in volts.                                                             | `device_mac`, `switch_id`              | `shelly_switch_voltage{device_mac="A8032ABE54DC",switch_id="switch:0"} 237.5`                 |
| `shelly_switch_aenergy_total` | Total accumulated energy of each switch in kWh.                                              | `device_mac`, `switch_id`              | `shelly_switch_aenergy_total{device_mac="A8032ABE54DC",switch_id="switch:0"} 6.532`           |
| `shelly_switch_temperature_c` | Temperature of each switch in Celsius.                                                      | `device_mac`, `switch_id`              | `shelly_switch_temperature_c{device_mac="A8032ABE54DC",switch_id="switch:0"} 23.5`            |
| `shelly_sys_uptime_seconds`   | Uptime of the device in seconds.                                                             | `device_mac`                           | `shelly_sys_uptime_seconds{device_mac="A8032ABE54DC"} 11081`                                  |
| `shelly_sys_ram_free`         | Free RAM in bytes.                                                                           | `device_mac`                           | `shelly_sys_ram_free{device_mac="A8032ABE54DC"} 151560`                                       |
| `shelly_sys_fs_free`          | Free filesystem space in bytes.                                                              | `device_mac`                           | `shelly_sys_fs_free{device_mac="A8032ABE54DC"} 180224`                                        |
| `shelly_wifi_rssi`            | WiFi signal strength in dBm.                                                                 | `device_mac`, `wifi_ssid`              | `shelly_wifi_rssi{device_mac="A8032ABE54DC",wifi_ssid="Brandmeldezentrale"} -54`              |
| `shelly_eth_ip`               | Indicates the Ethernet IP address of the device (label `eth_ip` holds the IP address).       | `device_mac`, `eth_ip`                 | `shelly_eth_ip{device_mac="A8032ABE54DC",eth_ip="10.33.55.170"} 1`                            |


## Development

### Project structure
```
shelly_exporter
├── cmd
│   └── exporter
│       └── main.go                # Entry point of the application
├── client
│   ├── api_client.go              # Handles API requests to Shelly endpoints
│   └── ShellyGetConfigResponse.go # API reponses typed in golang
├── metrics
│   └── metrics.go                 # Entry point to register the retrieval of metrics from the devices
├── config
│   └── config.go                  # Defines and mechanism to load config file
├── devices
│   ├── shelly.go                  # Defines basic types for all shelly devices
│   └── ShellySmartPlugS
│       └── main.go                # All go code related to getting metrics from a specific device type
├── go.mod                         # Go module dependencies
└── README.md                      # Documentation
```

### Adding a new device

## Contributing

We welcome contributions to this project! Feel free to:

* Open issues for bug reports or feature requests.
* Submit pull requests with enhancements or fixes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.