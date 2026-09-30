# Voice Gateway

Realtime voice gateway for 如此萌屋 · 芽系列·初芽 game consoles.

## Responsibilities

- authenticated WebSocket audio sessions
- audio framing, buffering, and playback
- speech recognition and synthesis adapters
- text conversation through `sub2api`
- child content policy and moderation
- latency, duration, and usage telemetry

## Run

```text
go run ./cmd/voice-gateway
```

The initial service exposes:

```text
GET /healthz
GET /readyz
```

## Configuration

Configuration is read from environment variables. See `configs/config.example.yaml`
for the intended deployment shape.
