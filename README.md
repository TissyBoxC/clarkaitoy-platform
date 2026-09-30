# Clarkaitoy

Clarkaitoy is a modular AI early-education companion platform for young children.
The workspace contains the parent application, admin console, backend services,
firmware, and shared contracts.

## Workspace

```text
apps/
  parent_app/          Flutter application for parents and guardians
  admin_web/           Vue 3 operations console
docs/                  Architecture and development documents
packages/
  contracts/           Versioned API, event, and capability contracts
services/
  device_platform/     Go control plane for accounts, devices, content, OTA
  voice_gateway/       Go realtime voice and AI gateway
  sub2api/             Independent AI gateway fork, managed separately
firmware/              ESP32-S3 N16R8 PlatformIO project
```

## Ownership Boundaries

- `parent_app` talks only to `device_platform`.
- `admin_web` talks to the platform management API.
- `voice_gateway` handles realtime audio, ASR, TTS, safety checks, and
  `sub2api` calls.
- `device_platform` owns child, family, device, content, OTA, and audit
  business data.
- `sub2api` owns model credentials, routing, quota, and upstream provider
  integrations. It must not contain child, family, device, or OTA data.
- `firmware` contains only device-side capabilities and device credentials.

## Repository Layout

`firmware/` and `services/sub2api/` are standalone repositories checked out
inside this workspace. They are intentionally ignored by the workspace
repository and must be committed in their own repositories.

Shared cross-service schemas live in `packages/contracts`. Service-local
transport contracts stay in each service's `contracts/` directory.

## Local Development

Each application and service is independently runnable:

```text
cd apps/parent_app && flutter run
cd apps/admin_web && npm run dev
cd services/device_platform && go run ./cmd/device-platform
cd services/voice_gateway && go run ./cmd/voice-gateway
cd firmware && platformio run -e esp32-s3-n16r8
```

Copy the corresponding `.env.example` or `config.example.yaml` file before
starting a service. Do not commit local credentials.

## Quality Gates

- Flutter: `flutter analyze` and `flutter test`.
- Vue: `npm run type-check`, `npm run build`, and `npm run test:e2e` when
  Playwright is installed.
- Go: `gofmt -w .`, `go test ./...`, and `go build ./...`.
- Firmware: `platformio run -e esp32-s3-n16r8`.

Every functional change must remain removable without breaking unrelated
modules. Commit one focused change set at a time with a descriptive message.
