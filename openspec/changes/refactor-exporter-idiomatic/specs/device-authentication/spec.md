## Purpose

Defines how the exporter authenticates to Shelly devices using credentials supplied in configuration.

## ADDED Requirements

### Requirement: Challenge-driven scheme selection
When a device is configured with credentials, the exporter SHALL determine the authentication scheme from the device's `WWW-Authenticate` challenge rather than assuming one, and SHALL authenticate with the advertised scheme.

#### Scenario: Device advertises Basic
- **WHEN** a device with configured credentials responds to an unauthenticated request with `401` and a `Basic` challenge
- **THEN** the exporter retries and sends an `Authorization` header using the Basic scheme with those credentials

#### Scenario: Device advertises Digest
- **WHEN** a device with configured credentials responds to an unauthenticated request with `401` and a `Digest` challenge
- **THEN** the exporter retries and sends an `Authorization` header using the Digest scheme

#### Scenario: Scheme reused after first challenge
- **WHEN** a device's scheme has been detected
- **THEN** subsequent requests to that device use the same scheme without re-probing for the challenge

### Requirement: Basic authentication
When a device advertises HTTP Basic authentication, the exporter SHALL send HTTP Basic credentials on requests to that device.

#### Scenario: Basic challenge answered
- **WHEN** a device advertises the Basic scheme and credentials are configured
- **THEN** requests to that device include an `Authorization` header using the Basic scheme with those credentials

### Requirement: Digest authentication
When a device advertises HTTP Digest authentication, the exporter SHALL compute a Digest response using the device's challenge and SHALL support the MD5 and SHA-256 algorithms.

#### Scenario: MD5 challenge
- **WHEN** a device advertises a Digest challenge using the MD5 algorithm
- **THEN** the exporter answers with a valid Digest `Authorization` header computed with MD5

#### Scenario: SHA-256 challenge
- **WHEN** a device advertises a Digest challenge using the SHA-256 algorithm
- **THEN** the exporter answers with a valid Digest `Authorization` header computed with SHA-256

#### Scenario: Unanswered digest challenge
- **WHEN** a device responds with HTTP 401 and the challenge cannot be satisfied
- **THEN** the exporter records the request as failed and continues without crashing

### Requirement: Unauthenticated devices
When a device is configured without credentials, the exporter SHALL send requests without an `Authorization` header.

#### Scenario: No credentials configured
- **WHEN** a device has no username or password configured
- **THEN** requests to that device contain no `Authorization` header

### Requirement: Credential confidentiality
The exporter SHALL NOT write device credentials to logs or error messages.

#### Scenario: Authentication failure logged
- **WHEN** a device rejects authenticated requests with an authorization error
- **THEN** the exporter logs the failure without including the username or password

### Requirement: Authentication failure handling
The exporter SHALL treat an authorization failure from a device as a request error and continue polling without crashing.

#### Scenario: Device rejects credentials
- **WHEN** a device responds to a request with HTTP 401
- **THEN** the exporter records the request as failed and continues with the remaining endpoints and devices
