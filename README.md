# 如此萌屋

English | [简体中文](README.zh-CN.md)

## 芽系列·初芽

“芽系列·初芽” is a modular AI early-education companion platform for young children.
The workspace contains the parent application, admin console, backend services,
firmware, and shared contracts.

## Workspace

This workspace uses a hybrid repository layout. First-party platform code is
versioned together, while firmware and the third-party AI gateway fork remain
independent repositories.

```text
sprout-platform/         this repository
  apps/
    parent_app/          Flutter application for parents and guardians
    admin_web/           Vue 3 operations console
  packages/
    contracts/           Versioned API, event, and capability contracts
  services/
    device_platform/     Go control plane for accounts, devices, content, OTA
    voice_gateway/       Go realtime voice and AI gateway
  tools/
    bootstrap.ps1        checks out the external repositories at locked commits
  workspace.lock.yaml    external repository revision lock

sprout-firmware/         separate repository, checked out as firmware/
sprout-sub2api-fork/     separate repository, checked out as services/sub2api_fork/
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

`firmware/` and `services/sub2api_fork/` are standalone repositories checked
out inside this workspace. They are intentionally ignored by the platform
repository and must be committed in their own repositories.

`workspace.lock.yaml` records the external repository revisions used by a
platform release. Run `tools/bootstrap.ps1` to clone or update those repositories
at the locked revisions.

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

## GitHub Repositories

Publish the platform repository first. After that, create and push the two
external repositories separately:

```text
platform:     sprout-platform
firmware:     sprout-firmware
sub2api fork: sprout-sub2api-fork
```

Keep the upstream remote on the `sub2api_fork` checkout for future rebases.
Do not push “芽系列·初芽” business code into the upstream repository.
