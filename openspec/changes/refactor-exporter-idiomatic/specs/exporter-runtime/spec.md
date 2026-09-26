## Purpose

Defines how the exporter process runs: the HTTP endpoints it serves, how the HTTP server is constructed and bounded, and how it shuts down.

## ADDED Requirements

### Requirement: Prometheus metrics endpoint
The exporter SHALL expose collected metrics in Prometheus exposition format at `GET /metrics`.

#### Scenario: Metrics are scraped
- **WHEN** a client sends `GET /metrics`
- **THEN** the response has HTTP status 200 and a body containing the registered Shelly metric families

### Requirement: Landing page
The exporter SHALL serve a minimal HTML landing page at `GET /` that links to the metrics endpoint.

#### Scenario: Root page is served
- **WHEN** a client sends `GET /`
- **THEN** the response has HTTP status 200 and contains a link to `/metrics`

### Requirement: Health endpoint
The exporter SHALL expose `GET /health` that reports process liveness via HTTP status, and SHALL NOT report a failure status because a device is unreachable or a poll cycle failed.

#### Scenario: Healthy process
- **WHEN** the process is running and a client sends `GET /health`
- **THEN** the response has HTTP status 200

#### Scenario: Device outages do not fail liveness
- **WHEN** one or more configured devices are unreachable or failing to poll
- **THEN** `GET /health` still returns HTTP status 200

### Requirement: Configurable listen address
The exporter SHALL bind the HTTP server to the address supplied as `listenAddress` in configuration.

#### Scenario: Custom listen address
- **WHEN** the configuration sets `listenAddress` to `:9090`
- **THEN** the exporter listens for HTTP requests on port 9090

### Requirement: Server timeouts
The HTTP server SHALL set read, write, and idle timeouts so slow or idle clients cannot hold connections indefinitely.

#### Scenario: Slow client connection
- **WHEN** a client connects but does not complete its request within the read timeout
- **THEN** the server closes the connection instead of waiting indefinitely

### Requirement: Graceful shutdown
The exporter SHALL stop accepting new requests and shut down the HTTP server upon receiving `SIGINT` or `SIGTERM`, and SHALL stop all device polling loops as part of shutdown.

#### Scenario: Termination signal received
- **WHEN** the process receives `SIGTERM`
- **THEN** it stops device polling loops, shuts the HTTP server down gracefully, and exits

#### Scenario: Shutdown failure is reported
- **WHEN** graceful shutdown returns an error
- **THEN** the exporter logs the error and exits with a non-zero status
