## Purpose

Defines how the exporter polls configured Shelly devices, selects which metrics to collect per device model, and behaves when a device or endpoint fails.

## ADDED Requirements

### Requirement: Periodic polling
The exporter SHALL poll each configured device repeatedly at the configured update interval and update that device's metrics on each successful poll.

#### Scenario: Repeated polling
- **WHEN** a device is registered and the update interval elapses
- **THEN** the exporter fetches the device's data and updates its metrics

#### Scenario: Interval respected
- **WHEN** the configured update interval is `30s`
- **THEN** consecutive polls for a device are separated by approximately 30 seconds

### Requirement: Per-model metric dispatch
The exporter SHALL select which metric families to collect based on the device model, and a newly supported model SHALL be added without modifying the collection logic for other models.

#### Scenario: Cover-capable model
- **WHEN** a registered device reports a cover-profile model
- **THEN** the exporter collects cover metrics and Wi-Fi status metrics for that device

#### Scenario: Switch-capable model
- **WHEN** a registered device reports a switch model
- **THEN** the exporter collects switch status, switch configuration, and Wi-Fi status metrics for that device

#### Scenario: Multi-channel switch model
- **WHEN** a registered device reports a switch model with more than one switch component (for example a `Pro4PM`)
- **THEN** the exporter collects switch status and switch configuration metrics for every switch component, each with its own `switch_id` label

#### Scenario: Dual-profile model
- **WHEN** a registered device reports the `Plus2PM` model with the `cover` profile
- **THEN** the exporter collects cover and Wi-Fi status metrics
- **WHEN** it reports the `switch` profile
- **THEN** the exporter collects switch status, switch configuration, and Wi-Fi status metrics

### Requirement: Component discovery
The exporter SHALL discover switch and cover component IDs from the device's `Shelly.GetConfig` response and collect metrics for each discovered component, so multi-channel devices are supported without per-model channel counts.

#### Scenario: Components enumerated
- **WHEN** `Shelly.GetConfig` contains components `switch:0` through `switch:3`
- **THEN** the exporter polls `Switch.GetStatus` and `Switch.GetConfig` for IDs 0, 1, 2, and 3

#### Scenario: No components discovered
- **WHEN** a device's configuration exposes no switch or cover component (for example a mocked response)
- **THEN** the exporter falls back to polling component ID 0

### Requirement: Absent fields omitted
The exporter SHALL omit a metric when the device response does not include the corresponding field, so it does not publish fabricated zero readings for unsupported capabilities.

#### Scenario: Relay without metering
- **WHEN** a switch status response contains only `output` and `temperature`
- **THEN** the exporter exposes switch state and temperature and does not expose power, voltage, current, frequency, or energy series for that switch

### Requirement: Unknown model handling
The exporter SHALL handle a device whose model is not recognized without panicking and SHALL log that no model-specific metrics are collected.

#### Scenario: Unrecognized model
- **WHEN** a registered device reports a model with no registered handler
- **THEN** the exporter continues polling generic metrics, logs the unrecognized model, and does not panic

### Requirement: Device registration
The exporter SHALL register a device by host and SHALL NOT start a second polling loop for a host that is already registered.

#### Scenario: Duplicate host
- **WHEN** a device host is registered twice
- **THEN** the exporter keeps a single polling loop for that host and logs that the device was already registered

### Requirement: Failure isolation
The exporter SHALL isolate failures so that an error polling or updating one device, or one RPC endpoint, does not stop polling loops for other devices or for subsequent intervals.

#### Scenario: One endpoint fails
- **WHEN** an RPC endpoint returns an error for a device
- **THEN** the exporter records the error and continues with the device's remaining endpoints and the next scheduled poll

#### Scenario: Device unreachable at startup
- **WHEN** a registered device is unreachable when the exporter starts
- **THEN** the exporter does not panic and continues to run for the other devices

### Requirement: Scrape success metric
The exporter SHALL expose a per-device metric named `shelly_up` with `device_mac` and `host` labels that reports whether the device's most recent poll cycle succeeded. The `host` label SHALL always be present; `device_mac` SHALL be reported when the device's MAC is known and MAY be the empty string until a device that is offline at startup reports its MAC.

#### Scenario: Successful poll
- **WHEN** a device's poll cycle completes successfully
- **THEN** its scrape-success metric is 1

#### Scenario: Failed poll
- **WHEN** a device's poll cycle fails
- **THEN** its scrape-success metric is 0 and the exporter continues polling

#### Scenario: MAC not yet known
- **WHEN** a device is registered but its MAC cannot be read because it is unreachable
- **THEN** its scrape-success metric is exposed with the `host` label and a `device_mac` of the empty string, and is set to 0

#### Scenario: MAC learned after recovery
- **WHEN** a device that previously had no known MAC later reports its device info
- **THEN** the exporter stops reporting the empty-`device_mac` series and reports the metric with the device's MAC

### Requirement: Metric identity
Every collected metric SHALL identify its source device by a stable device MAC label, and static device attributes SHALL be labels rather than process-global values. The `shelly_up` metric is the sole exception: its `device_mac` label is best-effort as defined in the scrape success metric requirement, and its `host` label always identifies the device.

#### Scenario: Multiple devices
- **WHEN** more than one device is configured and polled concurrently
- **THEN** each metric series carries the MAC of the device it came from and no device's values are attributed to another device

### Requirement: Safe Wi-Fi metrics
Wi-Fi metrics SHALL tolerate responses where the SSID or station IP is absent.

#### Scenario: Station IP absent
- **WHEN** a device status response omits the station IP or SSID
- **THEN** the exporter updates Wi-Fi metrics without dereferencing a nil value or panicking

### Requirement: Collector testability
Each metric collector SHALL be constructible against a supplied Prometheus registerer and SHALL update metrics without relying on package-level mutable state, so collectors can be instantiated more than once within a single process.

#### Scenario: Independent registries
- **WHEN** two collector instances are constructed against two separate registries in the same process
- **THEN** both construct successfully and each updates only its own metrics
