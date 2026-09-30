# Device Platform

Game console business service for Clarkaitoy.

## Responsibilities

- parent, family, and child accounts
- device registration and binding
- parent policies and content permissions
- content package metadata
- OTA release and upgrade tasks
- telemetry, audit, and notifications

## Run

```text
go run ./cmd/device-platform
```

The initial service exposes:

```text
GET /healthz
GET /readyz
```

## Configuration

Configuration is read from environment variables. See `configs/config.example.yaml`
for the intended deployment shape.
