## Purpose

Defines how the exporter loads, validates, and defaults its YAML configuration, including the semantics of the device update interval.

## ADDED Requirements

### Requirement: YAML configuration loading
The exporter SHALL read its configuration from a YAML file whose path is supplied on the command line via the `-config` flag, defaulting to `config.yaml` when the flag is omitted.

#### Scenario: Valid configuration
- **WHEN** the exporter is started with a path to a readable YAML file
- **THEN** it loads the configuration successfully

#### Scenario: Default path used
- **WHEN** the exporter is started without the `-config` flag
- **THEN** it reads its configuration from `config.yaml` in the working directory

#### Scenario: Missing configuration file
- **WHEN** the configured path does not exist or is a directory
- **THEN** the exporter fails to start and reports an error

### Requirement: Strict field handling
The exporter SHALL reject configuration files that contain unknown fields, so typos fail fast instead of being silently ignored.

#### Scenario: Unknown field present
- **WHEN** the configuration contains a field the schema does not define
- **THEN** the exporter fails to start and reports the offending field

### Requirement: Defaults
The exporter SHALL apply defaults for omitted optional settings: a default listen address and a default device update interval.

#### Scenario: Optional settings omitted
- **WHEN** the configuration omits `listenAddress` and `deviceUpdateInterval`
- **THEN** the exporter uses its documented default listen address and default update interval

### Requirement: Debug logging
The exporter SHALL accept a boolean `debug` setting that selects debug-level logging, defaulting to false when omitted.

#### Scenario: Debug enabled
- **WHEN** the configuration sets `debug` to `true`
- **THEN** the exporter emits debug-level log output

#### Scenario: Debug omitted
- **WHEN** the configuration omits `debug`
- **THEN** the exporter logs at info level and does not treat the omission as an error

### Requirement: Configuration validation
The exporter SHALL validate the loaded configuration and fail fast with a clear message when it is invalid.

#### Scenario: No devices configured
- **WHEN** the configuration contains no devices
- **THEN** the exporter fails to start and reports that at least one device is required

#### Scenario: Non-positive update interval
- **WHEN** the configured update interval is zero or negative
- **THEN** the exporter fails to start and reports an invalid interval

#### Scenario: Empty device host
- **WHEN** a configured device has an empty host
- **THEN** the exporter fails to start and reports the invalid device

### Requirement: Update interval format
The device update interval SHALL be expressed as a Go duration string such as `30s` or `1m`, and the exporter SHALL interpret it as the polling period directly.

#### Scenario: Duration string
- **WHEN** the configuration sets `deviceUpdateInterval` to `30s`
- **THEN** the polling period is 30 seconds

#### Scenario: Bare integer rejected
- **WHEN** the configuration sets `deviceUpdateInterval` to a bare integer such as `30`
- **THEN** the exporter fails to start and reports that a duration string is required
