# 平台仓库代码文档

## 1. 仓库信息

| 项目 | 内容 |
| --- | --- |
| 仓库 | `clarkaitoy-platform` |
| 本地路径 | `D:\service\clarkaitoy` |
| 默认分支 | `main` |
| 远端 | `https://github.com/TissyBoxC/clarkaitoy-platform.git` |
| 职责 | 家长端、管理端、业务服务、语音网关、共享契约和跨仓库文档 |

平台仓库不包含 `firmware` 和 `services/sub2api_fork` 的源码。两个外部仓库通过
[workspace.lock.yaml](../../../workspace.lock.yaml) 锁定版本，并由
[tools/bootstrap.ps1](../../../tools/bootstrap.ps1) 检出。

## 2. 组成

```text
apps/
  parent_app/              Flutter 家长控制端
  admin_web/               Vue 3 运营管理端
services/
  device_platform/         Go 设备与家庭业务平台
  voice_gateway/           Go 实时语音和 AI 网关
packages/
  contracts/               跨端 HTTP、事件、MQTT、capability 契约
docs/
  features/                功能范围和覆盖矩阵
  repositories/            各仓库代码文档
  *.md                     原有架构和实施文档
tools/
  bootstrap.ps1            外部仓库检出和锁定版本检查
```

## 3. 文档入口

| 子项目 | 文档 |
| --- | --- |
| 家长控制端 | [parent_app.md](parent_app.md) |
| 后台管理端 | [admin_web.md](admin_web.md) |
| 设备与家庭业务服务 | [device_platform.md](device_platform.md) |
| 实时语音服务 | [voice_gateway.md](voice_gateway.md) |
| 共享契约 | [contracts.md](contracts.md) |

## 4. 平台调用关系

```text
parent_app ───────────────┐
                          ├── HTTPS / WebSocket ──> device_platform
admin_web ────────────────┘                              │
                                                         ├── PostgreSQL
                                                         ├── Redis
                                                         └── MQTT/TLS

firmware ── MQTT/TLS ────────────────────────────────────> device_platform
    │
    └── WebSocket ──> voice_gateway ──> sub2api_fork ──> 模型供应商
                              │
                              ├── ASR 供应商
                              └── TTS 供应商
```

## 5. 模块化规则

平台仓库中的每个 feature 和业务模块都必须可以独立删除：

- 删除模块时，只允许修改组合根、路由表、依赖声明和模块注册表。
- 模块之间只通过公开 service、接口或契约交互。
- 禁止跨模块直接读取对方数据表、内部类型或私有文件。
- 跨模块 DTO 必须来自 `packages/contracts` 或服务公开接口。
- 每个模块必须包含自己的测试、配置项、错误码和迁移目录。
- 平台可以在部署配置中启用或禁用模块，但源码和测试基线仍然保留。

## 6. 功能覆盖

所有基础层、主流层、差异化层、P0、P1、P2 功能都在
[功能覆盖矩阵](../../features/feature-coverage.md) 中定义。平台仓库需要提供：

- 家长端、管理端和两个 Go 服务的用户入口。
- 儿童、家庭、设备、内容、策略、OTA、遥测、审计和通知闭环。
- ASR、LLM、TTS、内容安全、摄像头问答和用量统计闭环。
- 对 P2 能力的管理、配置、数据展示和设备控制能力。

## 7. 质量门禁

| 子项目 | 命令 |
| --- | --- |
| Flutter | `flutter analyze`、`flutter test` |
| Vue | `npm run type-check`、`npm run build`、`npm run test:e2e` |
| Go | `gofmt -w .`、`go test ./...`、`go build ./...` |
| 契约 | Schema 校验、兼容性检查、生成代码检查 |

## 8. 提交前检查

1. `git status --short --branch`
2. `git diff`
3. `git diff --cached`
4. 删除生成物、本地配置、无关文档和重复文件。
5. 确认 `AGENTS.md` 未进入提交。
6. 按 `<type>(<scope>): <summary>` 提交。
