## 1. Repository Layout

- [x] 1.1 Create `internal/` with `client`, `config`, `collector`, and `exporter` packages; move `config/` and `client/` into `internal/` with `git mv`. Verify `go build ./...` succeeds after updating imports.
- [x] 1.2 Consolidate the seven `rpc/*.*/` collector packages into `internal/collector` as one file per RPC family (`shelly_status.go`, `shelly_config.go`, `shelly_device_info.go`, `cover_status.go`, `switch_status.go`, `switch_config.go`, `wifi_status.go`), plus a new `up.go` for `shelly_up` (see 5.4), preserving metric names and labels. Verify `go build ./...` succeeds and no `rpc/` directory remains.
- [x] 1.3 Move the device manager and polling loop from `rpc/main.go` into `internal/exporter`, and update `cmd/exporter/main.go` imports. Verify `go build ./...` and `go vet ./...` are clean.
- [x] 1.4 Confirm no external importer depends on the directories moved under `internal/` (search the module for imports of the old non-`internal` paths, and check the module is not consumed as a library). Verify the only importers are within this repository.

## 2. Configuration

- [x] 2.1 Add `gopkg.in/yaml.v3` and remove `gopkg.in/yaml.v2` in `go.mod`; load config with `Decoder.KnownFields(true)`. Verify `go mod tidy` and a test that an unknown field is rejected.
- [x] 2.2 Introduce a `Duration` type whose `UnmarshalYAML` accepts a Go duration string via `time.ParseDuration` and rejects bare numbers; change `deviceUpdateInterval` to use it and remove the `* time.Second` multiplication. Verify table tests accept `30s`/`1m` and reject `30`.
- [x] 2.3 Add defaults (`listenAddress: :8080`, `deviceUpdateInterval: 30s`, `debug: false`) and validation (at least one device, non-empty host, positive interval). Verify tests cover each rejection path.
- [x] 2.4 Retain the `debug` flag (default `false`) and wire it to the logger level; verify a config containing `debug` loads under `KnownFields(true)` and that omitting it defaults to info-level logging.
- [x] 2.5 Change the `-config` flag default from `./config.yml` to `config.yaml` so an omitted flag matches the shipped example config, Dockerfile, README, and Helm configmap. Verify starting with no `-config` flag reads `config.yaml` and an explicit `-config` still overrides it.

## 3. HTTP Client and Authentication

- [x] 3.1 Define a `Fetcher` interface (`FetchData(ctx context.Context, endpoint string, result any) error`) in `internal/client` alongside `APIClient` (collectors keep importing `internal/client` for DTOs) and update collectors to depend on the interface rather than the concrete client; add `context` to the client request. Verify `go build ./...` compiles with the interface in place.
- [x] 3.2 Make `APIClient` accept an injectable `*http.Client` and optional credentials. Verify a test injects a custom client and observes the call.
- [x] 3.3 Add challenge-driven authentication: when credentials are configured, send the first request unauthenticated, parse the `WWW-Authenticate` challenge on `401`, and retry with Basic or Digest accordingly; cache the scheme (and digest challenge) per device so later requests do not re-probe. Verify `httptest` tests assert scheme selection, the `Authorization` header, absence of the header when no credentials are configured, and that credentials never appear in logs.
- [x] 3.4 Add `github.com/icholy/digest` and implement Digest responses supporting MD5 and SHA-256 with the MD5+SHA-256 test matrix. Verify digest tests assert a valid `Authorization` header for both algorithms and that an unsatisfiable challenge is recorded as a request error without crashing.
- [x] 3.5 Add client tests for non-200 responses, malformed JSON, 401 handling, and request timeouts/cancellation using `httptest.Server`. Verify `go test ./internal/client/...` passes.

## 4. Collectors

- [x] 4.1 Convert each collector to a struct constructed with `NewXxxCollector(reg prometheus.Registerer) (*XxxCollector, error)` using `reg.Register`, and remove all package-level `var metrics` and `MustRegister` on the default registry. Verify a test constructs two collectors against two `prometheus.NewRegistry()` instances without panic.
- [x] 4.2 Make Wi-Fi metric updates nil-safe for absent SSID/station IP. Verify a test with nil fields does not panic and still updates the remaining metrics.
- [x] 4.3 Add characterization tests per collector that update a struct from an `httptest` response and assert metric names, labels, and values with `prometheus/testutil`. Verify `go test ./internal/collector/...` passes.

## 5. Device State and Dispatch

- [x] 5.1 Store device model/profile/MAC on the per-device struct after reading device info, and re-attempt `Shelly.GetDeviceInfo` on each poll cycle until it succeeds so a device offline at registration resolves later; remove the `DeviceModel`/`DeviceProfile`/`DeviceMac` globals and the `GetDeviceType`/`GetDeviceProfile`/`GetDeviceMac` accessors. Verify `go test -race ./...` reports no data race with multiple devices.
- [x] 5.2 Handle device-info read failure at registration without dereferencing nil; verify a test with an unreachable device does not panic, continues polling generic handlers, and resolves the model-specific handler after device info becomes available.
- [x] 5.3 Replace the model `switch` with a `map[string]handler` dispatch registry and remove the duplicated PlusPlugS/Mini1G3 branch. Verify tests cover a cover model, a switch model, and an unknown model (logged, no panic).
- [x] 5.4 Add `internal/collector/up.go` exposing `shelly_up{device_mac,host}` via `NewUpCollector(reg prometheus.Registerer)` and inject it into the device manager; set it to 1 on a completed poll cycle and 0 on failure. Always set the `host` label; when the MAC is not yet known use the empty string, and call `DeleteLabelValues("", host)` once the device's MAC is learned so the empty-MAC series does not persist. Verify a test drives a success and a failure and asserts the gauge value for both labels, a test asserts `shelly_up{device_mac="",host}` is 0 for an unreachable device, and a test asserts the empty-MAC series disappears after device info is read.

## 6. Runtime Lifecycle

- [x] 6.1 Replace `DefaultServeMux` usage with `http.NewServeMux` and construct an `*http.Server` with read, write, and idle timeouts; keep `/`, `/metrics`, and `/health`. Verify handler unit tests assert `GET /`, `GET /health`, and `GET /metrics` status codes.
- [x] 6.2 Shut down on `SIGINT`/`SIGTERM` using `signal.NotifyContext`, cancelling device polling and calling `srv.Shutdown`. Verify a test cancels the context and observes polling loops stop.
- [x] 6.3 Replace `log.Fatal` + `slog` mixing with `slog` error logging and a non-zero exit; handle previously ignored errors (`w.Write`, `fmt.Fprint`) and fix the misleading/typo'd error wraps. Verify a code search shows no `log.Fatal` and `go vet ./...` is clean.
- [x] 6.4 Remove the always-true `serverIsHealthy()` stub and make `/health` return 200 for liveness without consulting device state. Verify a handler test asserts `/health` is 200 while a registered device is unreachable or a poll cycle has failed.

## 7. Naming and Cleanup

- [x] 7.1 Apply initialism naming (`MAC`, `SSID`, `RSSI`, `WiFi`), replace `interface{}` with `any`, remove the redundant `switch x := v; x` pattern, and avoid shadowing the `config` package. Verify `golangci-lint run` reports no `revive`/`golint` naming findings.
- [x] 7.2 Remove dead code: the unused `cfgPath` parameter and any other unused identifiers. Verify `deadcode`/`unused` linters are clean.

## 8. Tooling and Deployment Assets

- [x] 8.1 Add a `Taskfile.yml` with `build`, `test` (`go test -race ./...`), `lint`, `fmt`, `tidy`, `vet`, and `run` tasks. Verify each task runs successfully using the Task binary.
- [x] 8.2 Add a `.golangci.yml` enabling at least `govet`, `errcheck`, `staticcheck`, `revive`, `ineffassign`, and `unused`. Verify `golangci-lint run` passes.
- [x] 8.3 Add a GitHub Actions workflow that installs Task via `go-task/setup-task` and runs `task vet`, `task lint`, and `task test` (race detector) on push and pull requests. Verify the workflow passes on a branch.
- [x] 8.4 Add `AGENTS.md` documenting how to install Task (`go-task/setup-task` in CI; Homebrew or the install script for local development) and the `task build`, `task test`, and `task lint` commands. Verify it lists the tasks defined in 8.1.
- [x] 8.5 Update `config.yaml` and `charts/shelly-exporter/{values.yaml,README.md,templates/configmap.yaml}` to the duration-string format: change the chart `updateInterval` value from `30` (int seconds) to `30s` (duration string), update the README value table, and bump `Chart.yaml`. Verify the exporter starts with the updated example config and `helm template` renders `deviceUpdateInterval: 30s`.
- [x] 8.6 Sync the root `README.md` to the code: update the configuration section (duration-string `deviceUpdateInterval`, `username`/`password` authentication), document the new `shelly_up{device_mac,host}` metric, and correct every metric table to the names and labels the exporter actually exposes (for example `shelly_device_info`, not `device_info`). Verify each documented metric name matches its `prometheus.GaugeOpts` namespace/subsystem/name.
- [x] 8.7 Record the release: add a BREAKING entry to the changelog/release notes covering the `deviceUpdateInterval` format, the Helm `updateInterval` value type, and the `internal/` package-layout move, and bump `appVersion` (in `charts/shelly-exporter/Chart.yaml`). Verify the note names all three breaking changes and the version is incremented.

## 9. Verification

- [x] 9.1 Run `go test -race ./...` and confirm all tests pass with no races.
- [x] 9.2 Run `go build ./...`, `go vet ./...`, and `golangci-lint run` and confirm all are clean.
- [x] 9.3 Start the exporter against an `httptest`/mock device and confirm `/metrics` exposes the existing metric families plus the new scrape-success metric, and that `/health` returns 200.
- [x] 9.4 Run `openspec validate refactor-exporter-idiomatic --strict` and confirm the change validates.
