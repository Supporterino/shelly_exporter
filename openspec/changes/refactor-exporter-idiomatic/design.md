## Context

See `proposal.md` for motivation. The current implementation has seven collector packages (`rpc/Cover.GetStatus`, `rpc/Shelly.GetStatus`, ...) that each hold a package-level `var metrics` and call `prometheus.MustRegister` on the default registry, a `rpc` package with a per-device goroutine loop and a model `switch`, and a `client.APIClient` with no injection seam. `gopkg.in/yaml.v2` currently decodes `deviceUpdateInterval` as nanoseconds, which the code then multiplies by `time.Second`. The `openspec/specs/` directory is empty, so this change establishes the initial capability specs.

## Goals / Non-Goals

**Goals:**
- Make every collector and the API client constructible and testable in isolation, with no process-global mutable state.
- Fix the data race, startup nil dereference, nil Wi-Fi dereference, and duration footgun as part of the restructure.
- Establish an idiomatic Go layout and naming baseline that future device models and metrics can extend.

**Non-Goals:**
- Adding new metric families beyond the current set (the refactor must not change exposed metric names or label schemes unless required for correctness), except the scrape-success metric described below.
- Changing the Helm chart's public values beyond the config keys affected by the duration change.
- Introducing a metrics caching layer or changing scrape semantics.

## Decisions

### Decision: Single `internal/collector` package with one file per RPC family
Consolidate all collectors into `internal/collector` (files such as `shelly_status.go`, `cover_status.go`, `switch_config.go`, `wifi_status.go`). Keep `internal/client` (API client + response DTOs), `internal/config`, and `internal/exporter` (wiring + device manager + main).
**Why:** the seven packages differ only by their metric definitions and mapping; a single package removes the dotted/capitalized directory names, the seven import aliases, and the duplicated global/`MustRegister` boilerplate at once.
**Alternatives considered:** renaming each directory to lowercase while keeping them separate (preserves the boilerplate and aliases, no cohesion gain); keeping packages at repo root (loses the enforcement that these are not a public API).

### Decision: Inject `prometheus.Registerer` instead of `MustRegister`
Each collector exposes a constructor, e.g. `NewShellyStatusCollector(reg prometheus.Registerer) (*ShellyStatusCollector, error)`, which uses `reg.Register(...)` and returns the error. The exporter passes the default registerer; tests pass `prometheus.NewRegistry()`.
**Why:** `MustRegister` panics on duplicate registration, which makes a second construction in one test binary fatal. Returning an error also lets wiring fail cleanly. `MustRegister` can still be used at the single call site in the exporter if desired, but collectors themselves must not assume the default registry.
**Alternatives considered:** keeping `MustRegister` and relying on `prometheus.NewRegistry` everywhere (still panic-prone and hides errors); a `Register` method separate from the constructor (allows partially-constructed collectors).

### Decision: HTTP seam via interface + injectable credentials
Define a minimal `Fetcher` interface (`FetchData(ctx context.Context, endpoint string, result any) error`) in `internal/client`, alongside the concrete `APIClient` that implements it. Collectors accept a `Fetcher` for fetching but continue to import `internal/client` for the response DTOs they decode, so the interface sits with the DTOs it serves rather than in `internal/collector`. The API client holds an `*http.Client` (injectable), a base URL, and optional credentials. Authentication is applied by a challenge-driven `http.RoundTripper` installed on that client, keeping the `Fetcher` seam orthogonal to auth. Collectors depend on the interface, not the concrete client.
**Why:** `httptest.Server` plus the interface makes fetch behavior and collector mapping independently testable without a real device. Adding `context` enables per-request timeouts and shutdown cancellation, and the transport seam lets tests inject Basic or Digest challenges. Placing `Fetcher` with `APIClient` and the DTOs avoids a collector-package interface that collectors would implement against while still importing `internal/client`, which would be a half-seam.
**Alternatives considered:** defining `Fetcher` in `internal/collector` (the original design) keeps the interface near its consumers but is cosmetic while collectors must import `internal/client` for DTOs; splitting DTOs into a third package makes collectors depend only on DTOs + `Fetcher` but grows the diff for no behavioral gain.

### Decision: Challenge-driven Basic/Digest authentication
When a device is configured with credentials, the client sends its first request unauthenticated. On a `401`, it parses the `WWW-Authenticate` challenge and selects the scheme the device advertises: `Basic` (retry with `Authorization: Basic ...`) or `Digest` (compute credentials using `github.com/icholy/digest`, supporting MD5 and SHA-256). The selected scheme and the latest digest challenge are cached per device so steady-state requests authenticate without an extra round trip.
**Why:** Shelly Gen2 devices commonly advertise Digest rather than Basic, so a Basic-only client fails against them. Probing first avoids sending credentials (or the wrong scheme) to a device that will reject them, while caching keeps the steady-state cost at zero extra requests. Basic remains supported for devices that advertise it.
**Alternatives considered:** preemptive Basic (wrong for Digest-only devices and sends credentials before a challenge); a fixed scheme in configuration (shifts a protocol detail onto users); hand-rolled digest (more code to own and test than the small, MIT-licensed transport).

### Decision: Per-device state in a struct, no device globals
Attempt `Shelly.GetDeviceInfo` at registration and store the model/profile/MAC on the device's own struct. Remove `GetDeviceType`/`GetDeviceProfile`/`GetDeviceMac` and the `DeviceModel/Profile/Mac` package globals. If device-info is unavailable at registration, the device still registers and polls generic handlers; the poll loop re-attempts `Shelly.GetDeviceInfo` each cycle and resolves the model-specific handler (and the MAC label) once it succeeds.
**Why:** the globals are written by every device's poll loop and read during other devices' registration, which is the data race and cross-device misattribution. Per-device state removes both the race and the startup nil dereference (an unreachable device has no info; the handler is simply not resolved yet). Retrying instead of giving up keeps a device that was merely offline at startup observable and eventually fully collected.

### Decision: Handler registry for model dispatch
Replace the `switch` with `map[string]deviceHandler` (model to handler). A handler reflects a device's profile and runs the applicable collectors. A device with an unknown model gets the generic handlers only and a log line.
**Why:** removes the duplicated PlusPlugS/Mini1G3 branch and makes a new model a single map entry plus handler, without touching existing models (matches the spec requirement).

### Decision: Per-device scrape-success metric
Add a `shelly_up{device_mac,host}` gauge set to 1 when a poll cycle completes and 0 when it fails, owned by a new `internal/collector/up.go` collector (`NewUpCollector(reg prometheus.Registerer) (*UpCollector, error)`) that the device manager holds and updates. The update is driven from the device polling loop rather than from a per-RPC handler.
The `host` label is always known at registration, but `device_mac` is best-effort: a device that is offline when the exporter starts cannot report its MAC. In that case the loop emits `shelly_up{device_mac="",host="..."} 0` (so the failure is observable immediately), and once `Shelly.GetDeviceInfo` eventually succeeds the loop calls `DeleteLabelValues("", host)` and starts reporting the real MAC, avoiding a permanently stale empty-MAC series.
**Why:** today a failed poll is only visible in logs, so a device that goes offline silently keeps its last values; an up metric makes staleness visible to Prometheus and alerts. Poll-cycle scope matches the spec's failure-isolation behavior. Keeping `host` as the always-present identity means a device is observable even before its MAC is known, and the one-time series swap on recovery keeps the exposed set clean.
**Alternatives considered:** `host`-only labels (loses the MAC join and contradicts the spec); per-endpoint success metrics (more granular but higher cardinality and noisier for little operational gain); log-only (status quo, not alertable).

### Decision: Config as Go duration strings with a custom duration type
Use `gopkg.in/yaml.v3` with `Decoder.KnownFields(true)`, a `Duration` wrapper type whose `UnmarshalYAML` accepts a duration string parsed by `time.ParseDuration` and rejects bare numbers, plus defaults (`listenAddress: :8080`, `deviceUpdateInterval: 30s`, `debug: false`) and validation (at least one device, non-empty unique hosts, positive interval). The `debug` flag is retained (default `false`) to control log level so the shipped `config.yaml` and Helm configmap stay valid under `KnownFields(true)`. Remove the `* time.Second` multiplication.
**Why:** yaml decodes a bare `30` as nanoseconds, so the current code only works by coincidence and breaks catastrophically for `30s`; an explicit wrapper makes the accepted format unambiguous and testable. `KnownFields(true)` turns typos into startup errors.
**Alternatives considered:** staying on yaml.v2 and validating in code (no strict-field support, weaker); accepting both integer-seconds and duration strings (perpetuates the ambiguity the fix removes).

### Decision: Default configuration path is `config.yaml`
Change the `-config` flag default from `./config.yml` to `config.yaml` so an omitted flag resolves to the same file the repository, Dockerfile (`-config config.yaml`), README examples, and Helm configmap use. The explicit `-config` flag is unchanged.
**Why:** the current default points at a filename no shipped asset uses, so `./shelly_exporter` with no arguments fails on a correctly-installed directory. Aligning the default removes that footgun and makes every documented invocation work whether or not the flag is passed.
**Alternatives considered:** keeping `./config.yml` (still mismatched with all shipped assets); requiring `-config` with no default (a needless break for the documented `-config config.yaml` usage, which still works either way).

### Decision: HTTP server with timeouts and signal-driven graceful shutdown
Build an `*http.Server` with an explicit `http.NewServeMux`, set `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout`, and run it via a `signal.NotifyContext`-derived context so `SIGINT`/`SIGTERM` cancels the device manager and calls `srv.Shutdown`.
**Why:** avoids the global `DefaultServeMux`, removes a slow-client DoS vector, and ensures polling goroutines stop on shutdown.

### Decision: `/health` reports liveness only, not device health
`GET /health` returns 200 whenever the process is running; remove the always-true `serverIsHealthy()` stub and return 200 directly. It does not consult device or poll status.
**Why:** the endpoint is wired to the Helm liveness and readiness probes, so failing it on a single unreachable device would make Kubernetes restart a perfectly healthy exporter every time a device drops off the network. Device-level failure is deliberately surfaced through `shelly_up` instead, keeping liveness (process) and readiness (per-device) concerns separate.
**Alternatives considered:** reporting readiness from aggregate device state (ties pod restarts to device uptime and needs a policy for partial failure); a separate `/ready` endpoint (more surface than this change needs).

### Decision: Effective Go naming pass
Initialisms to all-caps (`MAC`, `SSID`, `RSSI`, `WiFi`), drop snake_case parameters (`device_mac` to `deviceMAC`), use `any` for `interface{}`, remove redundant `switch x := v; x`, avoid shadowing the `config` package, and remove dead code (`cfgPath` parameter, misleading/typo'd error wraps).

## Risks / Trade-offs

- **Digest adds protocol surface and MD5 is cryptographically weak** → rely on the small, maintained `github.com/icholy/digest` transport, support MD5 and SHA-256, cover both schemes with `httptest` challenge servers, and document that the scheme (and its weaknesses) is imposed by the device. Auth failures still degrade to logged request errors, not crashes.
- **Large, wide diff touches nearly every file** → phase tasks (A testability/behavior, B naming/layout, C runtime, D design) and keep each phase a separate reviewable commit so the change can be reverted in slices.
- **Config duration change is breaking** → update `config.yaml`, the Helm chart values/configmap, and release notes; keep the blast radius to one key with a clear startup error.
- **`internal/` move breaks any external importer** → the module is an exporter binary; verify no external consumers before merging, and treat the move as breaking in the changelog.
- **Refactor must not change metric names/labels** → cover with characterization tests using `prometheus/testutil` that assert current metric names, labels, and values before and after.

## Migration Plan

1. Land phases with tests green (`go test -race ./...`) and `golangci-lint` clean.
2. Update `config.yaml` and `charts/shelly-exporter/{values.yaml,README.md,templates/configmap.yaml}` to the duration-string form: the chart `updateInterval` value changes from `30` to `30s` and `Chart.yaml` is bumped.
3. Release with a BREAKING note for the `deviceUpdateInterval` format, the Helm `updateInterval` value type (int seconds to duration string), and package layout; users update configs from bare integers to `30s`.
4. Rollback: revert the change; no persisted state or external API is affected.

## Open Questions

<!-- None outstanding: the Digest-authentication question is resolved in favor of challenge-driven Basic/Digest authentication in this change (see Decisions). -->
