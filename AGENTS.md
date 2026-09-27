# AGENTS.md

Guidance for agents and contributors working on the Shelly Prometheus Exporter.

## Tooling

Development tasks are defined in `Taskfile.yml` and run with [Task](https://taskfile.dev).

### Installing Task

- **CI**: the GitHub Actions workflow installs Task with the
  [`go-task/setup-task`](https://github.com/go-task/setup-task) action.
- **Local development (Homebrew)**: `brew install go-task`
- **Local development (install script)**:
  `sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin`

### Tasks

| Task | Description |
|------|-------------|
| `task build` | Build the exporter binary (`bin/shelly_exporter`). |
| `task test` | Run all tests with the race detector (`go test -race ./...`). |
| `task lint` | Run `golangci-lint run`. |
| `task fmt` | Format all Go source files with `gofmt`. |
| `task tidy` | Tidy `go.mod` and `go.sum`. |
| `task vet` | Run `go vet ./...`. |
| `task run` | Run the exporter locally. |

Run `task build`, `task test`, and `task lint` before opening a pull request.

## Layout

- `cmd/exporter` — binary entry point.
- `internal/config` — YAML configuration loading, defaults, and validation.
- `internal/client` — Shelly RPC HTTP client, DTOs, and authentication.
- `internal/collector` — Prometheus collectors (one file per RPC family).
- `internal/exporter` — device manager, polling, HTTP server, and wiring.
