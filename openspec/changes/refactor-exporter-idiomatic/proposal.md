## Why

The exporter works, but its structure makes it both un-idiomatic Go and effectively untestable: seven packages hold package-global mutable metric state and register on the process-wide Prometheus registry with `MustRegister`, so collectors cannot be constructed twice or exercised in isolation. The same global state causes real data races and a startup nil-pointer panic, and there is currently no test suite, CI test job, or linting. This change makes the code idiomatic and test-backed in one coherent pass, because the testability problem and the style problem share the same root cause.

## What Changes

- Restructure packages into `internal/`: a single `collector` package (one file per RPC family), plus `client`, `config`, and `exporter`.
- **BREAKING** (Go API layout): remove all dotted/capitalized package directories (`Cover.GetStatus`, `Shelly.GetStatus`, etc.) and their import aliases; move packages under `internal/` so they are no longer importable externally.
- Convert package-global metric collectors into structs constructed with an injected `prometheus.Registerer`; remove `MustRegister` on the default registry and package-level `var metrics`.
- Introduce an HTTP seam (`HTTPDoer`/client interface) so collectors and the API client can be tested with `httptest`.
- Replace the per-model `switch` in `rpc/main.go` with a `map[model]handler` dispatch registry.
- Fix correctness defects: eliminate cross-device global state (data race), remove the startup nil dereference when device info is unavailable, guard nil-able Wi-Fi pointers, and fix misleading/typo'd error wraps.
- **BREAKING** (config): interpret `deviceUpdateInterval` as a Go duration string (`30s`); migrate config loading to `gopkg.in/yaml.v3` with `KnownFields(true)`, defaults, and validation, retaining the `debug` log-level flag so existing configs still load. Update `config.yaml` and the Helm chart accordingly.
- Wire `username`/`password` using HTTP Basic or Digest authentication, auto-detected from the device's `WWW-Authenticate` challenge (MD5 and SHA-256 supported).
- Harden the HTTP server: dedicated `http.NewServeMux`, read/write/idle timeouts, and graceful shutdown on `SIGINT`/`SIGTERM`.
- Add a per-device scrape-success metric so failed polls are observable in Prometheus instead of only in logs.
- Apply Effective Go naming: initialisms (`MAC`, `SSID`, `RSSI`), drop snake_case parameters, use `any` instead of `interface{}`.
- Add unit tests (stdlib + `prometheus/testutil`) with the race detector, a GitHub Actions test job, a `Taskfile.yml` (build, test, lint, fmt, tidy, vet, run), a `golangci-lint` config, and an `AGENTS.md`.

## Capabilities

### New Capabilities
- `exporter-runtime`: process lifecycle — metrics/health endpoints, HTTP server construction with timeouts, graceful shutdown.
- `device-metrics`: periodic device polling, per-model metric dispatch, collector construction, label semantics, and failure handling.
- `configuration`: YAML configuration loading — strict field handling, defaults, validation, and duration semantics.
- `device-authentication`: credential handling and Basic authentication of device requests.

### Modified Capabilities
<!-- None: openspec/specs/ is currently empty, so every capability is new. -->

## Impact

- Code: `cmd/exporter`, `config`, `client`, `metrics`, `rpc`, and all `rpc/*.*/*` packages; new `internal/` layout.
- Config/Deploy: `config.yaml`, `charts/shelly-exporter` (values + configmap), and any user configs using a bare integer interval.
- Dependencies: add `gopkg.in/yaml.v3` (and remove `gopkg.in/yaml.v2`); add `github.com/icholy/digest` for device authentication; add `prometheus/testutil` for tests; add `golangci-lint` in CI.
- Tooling: new `.github/workflows` test job (Task installed via `go-task/setup-task`), `Taskfile.yml`, `.golangci.yml`, `AGENTS.md`.
- Docs: root `README.md` synced to the code — config example (duration string, authentication), the new `shelly_up` metric, and metric name/label tables corrected to the names the exporter actually exposes.
- Release: a BREAKING changelog/release note and an `appVersion` bump recording the config format, Helm `updateInterval` type, and package-layout changes; verify no external importer depends on the directories moved under `internal/`.
- Behavior: devices with authentication now receive credentials; misconfigured configs fail fast instead of silently misbehaving; an offline or failed device is observable via `shelly_up`.
