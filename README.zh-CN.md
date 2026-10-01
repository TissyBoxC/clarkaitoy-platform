<div align="center">

<img src="assets/brand/sprout/brand_banner.png" alt="如此萌屋" width="720" />

# 如此萌屋

### 芽系列 · 初芽

面向幼儿的模块化 AI 早教陪伴平台

[English](README.md) | 简体中文

</div>

## 芽系列·初芽

芽系列·初芽面向 3 至 8 岁儿童及其监护人，覆盖游戏机本体、家长控制端、
后台管理端、业务服务、实时语音服务和共享契约。

项目按“功能可独立增加、独立删除”的原则组织。设备不支持摄像头、屏幕、触摸、
4G 或电池等硬件时，应能只移除对应模块，不修改无关业务。

> 所有能力默认关闭或采用最保守配置。社交、摄像头、麦克风上传和数据采集
> 必须由监护人主动开启，儿童隐私和安全优先于功能交付。

## 当前阶段

当前仓库处于工程骨架阶段，各项目已经建立可运行入口和模块边界，但大部分业务
尚未实现。

已经具备：

- `apps/parent_app`：可启动的 Flutter 应用骨架，包含路由、主题、配置边界和
  feature-first 目录。
- `apps/admin_web`：可构建的 Vue 3 管理端骨架，包含路由、状态容器、HTTP 边界
  和基础页面。
- `services/device_platform`：可测试、可构建的 Go 业务服务骨架，已划分账号、
  家庭、儿童、设备、家长策略、内容、OTA、遥测和审计等模块。
- `services/voice_gateway`：可测试、可构建的 Go 实时语音网关骨架，已划分音频、
  会话、ASR、TTS、LLM、内容安全和用量统计边界。
- `firmware`：独立的 PlatformIO 工程，目标芯片为 ESP32-S3 N16R8，构建目录为
  `bgen/`。
- `packages/contracts`：跨应用、跨服务的版本化契约目录。

## 系统组成

| 项目 | 面向对象 | 主要职责 |
| --- | --- | --- |
| `apps/parent_app` | 家长和监护人 | 登录、家庭、儿童档案、设备绑定、家长策略、使用报告、远程留言、内容和 OTA |
| `apps/admin_web` | 平台运营和管理人员 | 账号权限、家庭儿童设备管理、内容运营、AI 网关、OTA、审计和监控 |
| `services/device_platform` | 平台内部 | 家长与设备业务 API、设备通信、内容、家长策略、OTA、遥测、审计和通知 |
| `services/voice_gateway` | 平台内部 | 实时音频会话、ASR、TTS、内容安全、`sub2api` 调用和用量统计 |
| `services/sub2api_fork` | AI 基础设施 | 模型账号池、路由、配额、限流、计费和多供应商兼容接口 |
| `firmware` | 游戏机本体 | 唤醒、录音、播放、语音会话、内容缓存、网络、家长策略执行和 OTA |
| `packages/contracts` | 全项目共享 | HTTP、事件、MQTT 和设备 capability 的跨端契约 |

## 边界与调用关系

```text
家长 App ──────────────┐
                      ├──> device_platform ──> PostgreSQL / Redis / MQTT
后台 Web ──────────────┘

游戏机本体 ── MQTT/TLS ──> device_platform
     │
     └── WebSocket ─────> voice_gateway ──> sub2api ──> 模型供应商
                                  │
                                  ├──> ASR 服务
                                  └──> TTS 服务
```

主要边界：

- `parent_app` 只访问 `device_platform`，不直接连接 ESP32。
- `admin_web` 只访问平台管理 API，不在浏览器中保存供应商密钥。
- `device_platform` 负责儿童、家庭、设备、内容、家长策略、OTA 和审计数据。
- `voice_gateway` 负责实时语音和 AI 调用适配，不保存设备和家庭业务数据。
- `sub2api` 负责模型凭据、路由、配额和上游供应商，不包含儿童或设备业务。
- `firmware` 只保存设备身份、配置和必要缓存，不保存模型供应商密钥。

## 工作区结构

项目采用混合式仓库布局：平台代码放在同一个仓库，固件和 `sub2api` fork
保持独立仓库，并在平台工作区中以固定版本检出。

```text
sprout-platform/           当前平台仓库
  apps/
    parent_app/            Flutter 家长控制端
    admin_web/             Vue 3 后台管理端
  packages/
    contracts/             OpenAPI、事件、MQTT 和 capability 契约
  services/
    device_platform/       Go 设备与家庭业务服务
    voice_gateway/         Go 实时语音和 AI 网关
  tools/
    bootstrap.ps1          检出或更新锁定版本的外部仓库
  workspace.lock.yaml      外部仓库版本锁定

sprout-firmware/           独立固件仓库，本地路径为 firmware/
sprout-sub2api-fork/       独立 fork 仓库，本地路径为 services/sub2api_fork/
```

`firmware/` 和 `services/sub2api_fork/` 由平台根目录的 `.gitignore` 排除，
需要分别在各自仓库中提交和推送。`workspace.lock.yaml` 只记录平台发布所引用的
外部仓库版本，不把它们变成 Git submodule。

## 技术栈

| 层级 | 当前或推荐技术 |
| --- | --- |
| 家长端 | Flutter、Dart、go_router；后续可接入 Riverpod 或 Bloc、dio、freezed |
| 管理端 | Vue 3、TypeScript、Vite、Pinia、Vue Router、Axios |
| 业务服务 | Go 1.27.1；推荐 Gin 或 Echo、PostgreSQL、Redis、MQTT/TLS |
| 实时语音 | Go、WebSocket、Opus、ASR/TTS 适配器、`sub2api` |
| 管理数据库 | PostgreSQL 16 |
| 缓存与消息 | Redis 7，后续可选 NATS JetStream 或 Redis Streams |
| 设备通信 | MQTT/TLS、WebSocket、HTTPS |
| 固件 | ESP-IDF、C、FreeRTOS、PlatformIO |
| 目标硬件 | ESP32-S3 N16R8，16 MB Flash，8 MB OPI PSRAM |
| 固件构建目录 | `bgen/` |
| API 契约 | OpenAPI 3；异步消息预留 AsyncAPI 或事件 Schema |
| 观测 | OpenTelemetry、Prometheus、Grafana、Loki、Jaeger 或 Tempo |

## 环境准备

根据要开发的项目安装以下工具：

- Git
- Flutter，满足 Dart SDK 3.11.5 及以上
- Node.js 22.18 或 24.12 及以上，以及 npm
- Go 1.27.1 及以上
- PlatformIO CLI，仅在开发固件时需要
- Docker，后续用于 PostgreSQL、Redis、MQTT 和其他本地依赖

首次检出平台仓库后，可在平台根目录执行：

```powershell
.\tools\bootstrap.ps1
```

脚本会在缺少目录时克隆 `firmware` 和 `services/sub2api_fork`，已经存在时保留
当前检出。需要拉取远端信息时使用：

```powershell
.\tools\bootstrap.ps1 -Update
```

外部仓库必须位于 `workspace.lock.yaml` 记录的版本。`sub2api_fork` 同时保留
`origin` 和 `upstream`：`origin` 是可修改的 `sprout-sub2api-fork`，
`upstream` 只用于同步上游代码，禁止向 `upstream` 推送“芽系列·初芽”业务。

## 本地运行

各应用和服务可以独立启动。

### 家长控制端

```powershell
Set-Location apps/parent_app
flutter pub get
flutter run
```

### 后台管理端

```powershell
Set-Location apps/admin_web
npm install
npm run dev
```

### 业务服务

```powershell
Set-Location services/device_platform
go run ./cmd/device-platform
```

当前服务直接读取环境变量，常用变量如下。完整部署形态参考
`configs/config.example.yaml`。

```powershell
$env:DEVICE_PLATFORM_HTTP_HOST = "0.0.0.0"
$env:DEVICE_PLATFORM_HTTP_PORT = "8081"
$env:DEVICE_PLATFORM_DATABASE_DSN = "postgres://device_platform:change-me@127.0.0.1:5432/device_platform?sslmode=disable"
$env:DEVICE_PLATFORM_REDIS_ADDRESS = "127.0.0.1:6379"
$env:DEVICE_PLATFORM_MQTT_BROKER = "tcp://127.0.0.1:1883"
go run ./cmd/device-platform
```

### 实时语音服务

```powershell
Set-Location services/voice_gateway
go run ./cmd/voice-gateway
```

当前服务直接读取环境变量，常用变量如下。完整部署形态参考
`configs/config.example.yaml`。

```powershell
$env:VOICE_GATEWAY_HTTP_HOST = "0.0.0.0"
$env:VOICE_GATEWAY_HTTP_PORT = "8082"
$env:VOICE_GATEWAY_REDIS_ADDRESS = "127.0.0.1:6379"
$env:VOICE_GATEWAY_SUB2API_BASE_URL = "http://127.0.0.1:8080"
$env:VOICE_GATEWAY_SUB2API_API_KEY = "<local-api-key>"
go run ./cmd/voice-gateway
```

两个 Go 服务当前都提供健康检查：

```text
GET /healthz
GET /readyz
```

### 游戏机本体

```powershell
Set-Location firmware
platformio run -e esp32-s3-n16r8
```

固件的所有构建中间文件、镜像和测试输出统一写入 `bgen/`，该目录不进入 Git。
烧录和串口监视仍使用 PlatformIO 或 ESP-IDF 命令。

## 配置与凭据

- 本地凭据不得提交到 Git。
- 使用各项目提供的 `.env.example` 或 `config.example.yaml` 作为配置参考。
- 当前 Go 服务从环境变量加载配置，`config.example.yaml` 只用于说明部署形态。
- 不要把 `.env`、设备证书、私钥、模型供应商密钥写入源码或共享契约。
- 固件只保存设备身份和设备令牌，不保存模型供应商凭据。
- 管理端只能显示脱敏后的配置，不向浏览器返回真实密钥。
- 提交前检查根目录和各独立仓库的忽略规则是否覆盖本地生成文件。

## 质量门禁

### Flutter

```powershell
Set-Location apps/parent_app
flutter analyze
flutter test
```

### Vue 管理端

```powershell
Set-Location apps/admin_web
npm run type-check
npm run build
npm run test:e2e
```

首次运行端到端测试前，需要安装 Playwright 浏览器：

```powershell
npx playwright install
```

### Go 服务

```powershell
Set-Location services/device_platform
gofmt -w .
go test ./...
go build ./...

Set-Location ../voice_gateway
gofmt -w .
go test ./...
go build ./...
```

### 固件

```powershell
Set-Location firmware
platformio run -e esp32-s3-n16r8
```

每个功能修改都应补对应测试，并至少覆盖受影响项目的质量门禁。删除模块后，
还要验证无关模块仍能独立编译和启动。

## 模块化约束

模块边界是本项目的一等约束，不能只在目录上做表面拆分。

- 包名、模块名、方法名和变量名必须表达业务含义，并保持项目内一致。
- 公共 API 和复杂流程必须有面向开发者的简洁注释，说明参数、失败行为和约束。
- 模块之间只通过公开接口或 service 调用，禁止跨模块直接访问对方数据表。
- 单个功能模块必须能独立删除，不需要修改无关业务代码。
- 固件可选硬件必须使用独立 Kconfig 开关；关闭后，其初始化、依赖和资源都应被
  条件编译移除。
- 跨端协议必须进入 `packages/contracts`，接口变更必须带 `schema_version`。
- 服务内部传输细节留在各自的 `contracts/` 或 `internal/contracts/`，不要污染
  全项目共享契约。

## Git 工作流

仓库按以下边界独立管理：

| 仓库 | 本地路径 | 用途 |
| --- | --- | --- |
| `sprout-platform` | `D:\service\clarkaitoy` | 应用、服务、共享契约和文档 |
| `sprout-firmware` | `D:\service\clarkaitoy\firmware` | ESP32-S3 固件 |
| `sprout-sub2api-fork` | `D:\service\clarkaitoy\services\sub2api_fork` | `sub2api` fork |

提交信息统一使用：

```text
<type>(<scope>): <summary>
```

常用类型：

```text
feat
fix
refactor
perf
test
docs
build
chore
```

示例：

```text
feat(device_platform): add device activation flow
fix(voice_gateway): reject expired websocket sessions
docs(workspace): update external repository revisions
```

每次修改前，先确认对应仓库工作区干净且已有提交作为回归基线。每个独立功能完成
后，由开发者或执行修改的 AI 生成一次聚焦的 commit。修改外部仓库时必须在对应
仓库内提交和推送，不在平台仓库中混入其源码。

## 近期重点

1. 固化身份、家庭、儿童、设备和 capability 契约。
2. 完成 `device_platform` 的数据库、认证和设备绑定闭环。
3. 完成 `voice_gateway` 的 WebSocket、ASR、`sub2api` 和 TTS 闭环。
4. 接入家长端的登录、设备绑定、儿童档案和家长策略。
5. 接入管理端的设备、内容、OTA 和审计页面。
6. 按 Kconfig 开关逐项启用固件音频、网络、策略和 OTA 能力。
7. 补充 CI、监控、压测、隐私合规、灰度发布和灾难恢复。

## 儿童隐私与安全

- 所有公网通信使用 TLS。
- 儿童语音、照片、位置等数据按最小必要原则采集和处理。
- 未获得明确授权时，不保存儿童原始语音和图像。
- 日志必须脱敏，不记录密码、设备密钥、完整音频或真实身份信息。
- 模型请求必须经过内容安全策略，设备和管理端不持有模型供应商密钥。
- 管理操作、敏感数据访问、设备凭证变更和 OTA 发布必须留存审计记录。
- 提供家庭数据导出、删除和账号注销流程。
