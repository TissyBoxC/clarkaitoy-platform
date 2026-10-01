<div align="center">

<img src="assets/brand/sprout/brand_avatar.png" alt="如此萌屋" width="180" />

# 如此萌屋

### 芽系列 · 初芽

面向幼儿的模块化 AI 早教陪伴平台

English | [简体中文](README.zh-CN.md)

[![Go](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Flutter](https://img.shields.io/badge/Flutter-Dart%203.11-02569B?logo=flutter&logoColor=white)](https://flutter.dev/)
[![Vue 3](https://img.shields.io/badge/Vue%203-TypeScript-4FC08D?logo=vuedotjs&logoColor=white)](https://vuejs.org/)
[![ESP-IDF](https://img.shields.io/badge/ESP--IDF-ESP32--S3-E7352C?logo=espressif&logoColor=white)](https://docs.espressif.com/projects/esp-idf/en/stable/esp32s3/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)](https://redis.io/)

<!-- COMMUNITY_LINKS_START: 群链接待补充。 -->
<!-- DOCUMENTATION_LINKS_START: 项目文档入口待补充。 -->

</div>

> `sprout-platform` is the platform repository for 如此萌屋's first product,
> 芽系列·初芽. It contains the parent application, admin console, backend
> services, shared contracts, and workspace tooling. Firmware and the AI
> gateway fork remain independent repositories.

## 芽系列·初芽

芽系列·初芽面向 3 至 8 岁儿童及其监护人，目标是把游戏机本体、家长控制端、
后台管理端和 AI 服务组合成一套可持续演进的产品。平台默认保护儿童隐私，
所有摄像头、麦克风、社交和数据采集能力默认关闭，需由监护人明确开启。

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

## Technology

| Layer | Technology | Documentation |
| --- | --- | --- |
| Parent app | Flutter, Dart, Riverpod, go_router, Dio, secure storage | [Flutter](https://docs.flutter.dev/) · [Dart](https://dart.dev/guides) · [Riverpod](https://riverpod.dev/) · [go_router](https://pub.dev/packages/go_router) · [Dio](https://pub.dev/packages/dio) |
| Admin console | Vue 3, TypeScript, Vite, Pinia, Vue Router, Axios | [Vue 3](https://vuejs.org/guide/) · [TypeScript](https://www.typescriptlang.org/docs/) · [Vite](https://vite.dev/guide/) · [Pinia](https://pinia.vuejs.org/) · [Vue Router](https://router.vuejs.org/) · [Axios](https://axios-http.com/docs/intro) |
| Device platform | Go, PostgreSQL, Redis, MQTT/TLS | [Go](https://go.dev/doc/) · [PostgreSQL](https://www.postgresql.org/docs/) · [Redis](https://redis.io/docs/latest/) · [Eclipse Mosquitto](https://mosquitto.org/documentation/) |
| Voice gateway | Go, WebSocket, Opus, ASR, TTS, Sub2API | [Go](https://go.dev/doc/) · [WebSocket](https://datatracker.ietf.org/doc/html/rfc6455) · [Opus](https://opus-codec.org/docs/) |
| Firmware | ESP-IDF, C, FreeRTOS, PlatformIO | [ESP-IDF](https://docs.espressif.com/projects/esp-idf/en/stable/esp32s3/) · [FreeRTOS](https://www.freertos.org/Documentation/RTOS_book.html) · [PlatformIO](https://docs.platformio.org/) |
| Local services | Docker Compose, PostgreSQL, Redis, MQTT/TLS | [Docker Compose](https://docs.docker.com/compose/) · [PostgreSQL](https://www.postgresql.org/docs/) · [Redis](https://redis.io/docs/latest/) · [Eclipse Mosquitto](https://mosquitto.org/documentation/) |

## Open Source Notices

This repository uses third-party open-source software. Each dependency remains
under its own license; the project does not relicense third-party code.

- [Go standard library and toolchain](https://go.dev/LICENSE) - BSD-3-Clause.
- [Flutter and Dart](https://github.com/flutter/flutter/blob/master/LICENSE) - BSD-3-Clause.
- [Vue 3](https://github.com/vuejs/core/blob/main/LICENSE) - MIT.
- [Vite](https://github.com/vitejs/vite/blob/main/LICENSE) - MIT.
- [Pinia](https://github.com/vuejs/pinia/blob/v3/LICENSE) - MIT.
- [Vue Router](https://github.com/vuejs/router/blob/main/LICENSE) - MIT.
- [Axios](https://github.com/axios/axios/blob/v1.x/LICENSE) - MIT.
- [Dio](https://github.com/cfug/dio/blob/main/dio/LICENSE) - MIT.
- [Riverpod](https://github.com/rrousselGit/riverpod/blob/master/LICENSE) - MIT.
- [go_router](https://github.com/flutter/packages/blob/main/packages/go_router/LICENSE) - BSD-3-Clause.
- [PostgreSQL](https://www.postgresql.org/about/licence/) - PostgreSQL License.
- [Redis](https://github.com/redis/redis/blob/unstable/LICENSE.txt) - AGPL-3.0 or RSALv2/SSPLv1 depending on distribution.
- [Eclipse Mosquitto](https://github.com/eclipse-mosquitto/mosquitto/blob/master/LICENSE.txt) - EPL-2.0.
- [ESP-IDF](https://github.com/espressif/esp-idf/blob/master/LICENSE) - Apache-2.0.
- [FreeRTOS Kernel](https://github.com/FreeRTOS/FreeRTOS-Kernel/blob/main/LICENSE.md) - MIT.
- [PlatformIO Core](https://github.com/platformio/platformio-core/blob/develop/LICENSE) - Apache-2.0.

The platform repository does not yet declare a project-level license. No
license is granted for first-party code until a `LICENSE` file is added.

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
