# Shelly Prometheus Exporter

Shelly Prometheus Exporter is a Go-based application that collects metrics from Shelly devices via their REST API endpoints and exposes them in Prometheus-compatible format. This tool enables detailed monitoring and visualization of your Shelly devices in Prometheus and Grafana.

## Usage

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