# sub2api fork 代码文档

## 1. 仓库信息

| 项目 | 内容 |
| --- | --- |
| 仓库 | `clarkaitoy-sub2api-fork` |
| 本地路径 | `D:\service\clarkaitoy\services\sub2api_fork` |
| 上游 | `https://github.com/Wei-Shaw/sub2api.git` |
| fork | `https://github.com/TissyBoxC/clarkaitoy-sub2api-fork.git` |
| 许可证 | LGPL-3.0 或更高版本 |
| 后端 | Go 1.27 + Gin + Ent + Wire + Viper |
| 前端 | Vue 3 + Pinia + Vue Router + Tailwind + Chart.js |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |

## 2. 定位

`sub2api_fork` 是 Clarkaitoy 的 AI API 网关基础设施，负责模型账号池、供应商
适配、路由、配额、限流、计费、用量统计和调用审计。

它不保存儿童档案、家庭关系、设备生命周期、OTA 和家长策略。Clarkaitoy 业务
通过 `device_platform` 和 `voice_gateway` 的适配层调用它。

## 3. 魔改边界

允许的 Clarkaitoy 扩展：

- 组织、项目、租户或设备维度的请求标签。
- 儿童设备模型白名单和内容安全策略。
- 调用来源、设备、会话和用途审计字段。
- 超时、重试、熔断、降级和故障转移策略。
- 面向平台服务的内部管理 API。
- 统一调用日志、成本、用量和错误格式。

禁止写入：

- 儿童真实姓名、生日、头像和家庭关系。
- 设备配网、MQTT、GPIO、OTA 和固件逻辑。
- 家长时长策略、内容运营和课程体系。
- 设备长期密钥和固件签名密钥。

## 4. 目录概览

```text
services/sub2api_fork/
  backend/
    cmd/server/            服务入口
    ent/schema/            Ent schema
    internal/
      config/              配置加载
      domain/              领域模型
      handler/             HTTP handler
      middleware/          鉴权、限流、审计中间件
      model/               数据模型
      repository/          数据访问
      server/              路由和服务器
      service/             业务逻辑
      web/                 后端内嵌前端资源
    migrations/            SQL 迁移
    pkg/pluginapi/         插件 API
  frontend/
    src/
      api/                 API 客户端
      components/          通用组件
      features/            功能模块
      router/              路由
      stores/              Pinia store
      types/               TypeScript 类型
      views/               页面
      i18n/                国际化
  deploy/                  Docker Compose、配置和安装脚本
  docs/                    上游文档
  openspec/                规格和变更记录
  tools/                   构建和运维工具
```

## 5. 仓库职责

| 子项目 | 文档 | 职责 |
| --- | --- | --- |
| Go 后端 | [backend.md](backend.md) | 模型路由、账号池、配额、计费、审计和管理 API |
| Vue 管理端 | [admin-frontend.md](admin-frontend.md) | 账号、渠道、模型、配额、支付、审计和系统设置 |

## 6. 与 Clarkaitoy 的集成

详见 `device_platform` 文档中的 `ai_gateway` 模块和 `voice_gateway` 的
`llm/sub2api_client`。

建议请求标签：

```text
clarkaitoy.tenant_id
clarkaitoy.family_id
clarkaitoy.device_id
clarkaitoy.child_age_tier
clarkaitoy.capability_profile
clarkaitoy.policy_version
clarkaitoy.request_source
clarkaitoy.session_id
```

其中儿童标识只使用不可逆的内部标识或最小化标识，不传递真实身份信息。

## 7. 维护方式

- `origin` 只推送 Clarkaitoy 的 fork 修改。
- `upstream` 只用于 fetch 和 rebase，禁止 push。
- 魔改应投入独立分支或隔离目录，减少上游 rebase 冲突。
- 保留许可证、版权声明和修改源码。
- 每次同步上游后，重新验证 Clarkaitoy 扩展和契约测试。

## 8. 质量门禁

```powershell
Set-Location backend
go test -tags=unit ./...
go test -tags=integration ./...
golangci-lint run ./...
go generate ./ent
go generate ./cmd/server

Set-Location ../frontend
pnpm install --frozen-lockfile
pnpm run lint:check
pnpm run typecheck
pnpm run test:run
pnpm run build
```

## 9. 提交前检查

1. 确认当前分支、上游和远端正确。
2. 确认 `pnpm-lock.yaml`、Ent 生成代码和 Wire 生成代码同步。
3. 检查 `git status`、`git diff` 和 `git diff --cached`。
4. 不提交 Clarkaitoy 真实儿童、家庭或设备数据。
5. 不提交 `AGENTS.md`、构建产物、日志和本地配置。
