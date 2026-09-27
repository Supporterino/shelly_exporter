# Changelog

## 0.3.0

### Breaking changes

- **Configuration: `deviceUpdateInterval` is now a Go duration string.** Set it to
  a value such as `30s` or `1m` instead of a bare integer. Bare integer values
  are rejected at startup. Unknown configuration fields are now rejected as well.
- **Helm chart: `shellyexporter.updateInterval` changed type from `int`
  (seconds) to a duration `string`.** Update values such as `updateInterval: 30`
  to `updateInterval: 30s`.
- **Go package layout:** packages moved under `internal/`
  (`internal/config`, `internal/client`, `internal/collector`,
  `internal/exporter`) and are no longer importable as a public API. The dotted
  `rpc/*` collector packages were consolidated into `internal/collector`.
- **Metric labels:** `shelly_wifi_rssi` no longer carries an `rssi` label (the
  RSSI remains the sample value). Switch and cover metric families now carry a
  `name` label with the configured channel name. Update dashboards and alerts
  that select on these label sets.

### Added

- `shelly_up{device_mac,host}` scrape-success metric per device.
- Challenge-driven HTTP Basic / Digest (MD5 and SHA-256) device authentication
  from configured `username`/`password`.
- Graceful shutdown on `SIGINT`/`SIGTERM` and HTTP server timeouts.
- Unit tests, race-detector test runs, `golangci-lint`, a `Taskfile.yml`, and CI.
