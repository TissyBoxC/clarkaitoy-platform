# sub2api 后端代码文档

## 1. 定位

sub2api 后端是 AI API 网关，负责账号池、渠道、模型、路由、配额、限流、计费、
调用日志、插件和后台管理 API。

## 2. 技术栈

| 层级 | 当前实现 |
| --- | --- |
| 语言 | Go 1.27.0 |
| HTTP | Gin 1.9.1 |
| ORM | Ent 0.14.5 |
| 依赖注入 | Google Wire 0.7.0 |
| 配置 | Viper 1.18.2 |
| 日志 | `log/slog` + Zap + Lumberjack |
| 数据库 | PostgreSQL 15+ |
| 缓存 | Redis 7+ |
| JWT | `golang-jwt/jwt/v5` |
| WebSocket | Coder WebSocket + Gorilla WebSocket |
| WebAuthn | `go-webauthn/webauthn` |
| 测试 | Testify + Testcontainers + `go-sqlmock` |
| 任务调度 | `robfig/cron/v3` + Pond |
| 对象存储 | AWS SDK for Go v2 / S3 兼容 |

## 3. 目录结构

```text
backend/
  cmd/
    server/                 主服务
    cleanup-ingress-reject-logs/
    jwtgen/
    profit-preview/
  ent/
    schema/                 Ent schema
    migrate/                迁移
  internal/
    config/                 配置和 Wire
    domain/                 领域模型
    handler/                HTTP handler
    integration/            外部集成
    middleware/             中间件
    model/                  数据模型
    payment/                支付
    pkg/                    内部通用包
    platform/               平台能力
    repository/             数据访问
    securityaudit/          安全审计
    server/                 路由和服务器
    service/                业务逻辑
    setup/                  初始化
    util/                   工具
    web/                    内嵌前端
  migrations/               SQL 迁移
  resources/                模型定价等资源
  pkg/pluginapi/            插件 API
```

## 4. 核心模块

| 模块 | 职责 |
| --- | --- |
| `auth` | 登录、注册、OAuth、Passkey、JWT、刷新令牌、MFA |
| `account` | 模型供应商账号、凭据、状态、平台和分组 |
| `group` | 账号组、用户组、模型白名单和权限 |
| `channel` | 渠道、定价、监控和上游状态 |
| `gateway` | OpenAI、Anthropic、Gemini、Grok 等协议转换和转发 |
| `routing` | 模型路由、调度、故障转移、粘性和熔断 |
| `billing` | 余额、订阅、配额、预留、计费和使用日志 |
| `payment` | Stripe、Airwallex、微信、支付宝等支付适配 |
| `usage` | Token、图片、视频、音频和调用统计 |
| `risk_control` | 风控、Prompt 审计和安全策略 |
| `plugin` | 插件加载、RPC 和请求钩子 |
| `admin` | 后台用户、账号、渠道、系统设置和审计 API |
| `ops` | 监控、告警、清理任务和运行时诊断 |

## 5. Clarkaitoy 扩展建议

### 5.1 请求上下文

建议在请求上下文中增加：

```go
type ClarkaitoyContext struct {
    TenantID        string
    FamilyID        string
    DeviceID        string
    ChildAgeTier    string
    Capability      string
    PolicyVersion   string
    RequestSource   string
    SessionID       string
}
```

字段只用于路由、配额、审计和安全策略。真实儿童身份和家庭信息保留在
`device_platform`。

### 5.2 模型策略

- 根据设备 capability、年龄和套餐选择模型白名单。
- 对幼儿对话限制昂贵模型、长上下文和高风险能力。
- 将语音、图片、视频和文本请求分别记录用量。
- 供应商失败时按策略降级或明确返回错误。
- 对不合规请求返回稳定错误码，不暴露供应商原始错误。

### 5.3 审计

每次调用记录：

```text
request_id
trace_id
tenant_id
device_id
request_source
model
provider
input_tokens
output_tokens
image_count
audio_seconds
cost
latency_ms
status
error_code
policy_version
```

不记录儿童真实姓名、完整音频、图片内容、提示词原文或供应商密钥。

## 6. 配置

- 主配置由 Viper 读取 YAML 和环境变量。
- 密钥、数据库密码、Redis 密码和支付凭据只从环境变量或 Secret 管理加载。
- Clarkaitoy 扩展配置使用独立前缀，避免与上游键冲突。
- 配置变更必须记录版本和审计。

## 7. 数据库和迁移

- Ent schema 是模型定义源。
- 修改 schema 后运行 `go generate ./ent`。
- 数据库迁移必须可回滚或在发布说明中明确不可逆变更。
- Clarkaitoy 扩展表使用独立前缀，例如 `clarkaitoy_*`。
- 不把儿童、家庭和设备业务表复制进 sub2api 数据库。

## 8. API 边界

| API | 用途 |
| --- | --- |
| `/v1/*` | OpenAI 兼容模型接口 |
| `/v1/messages` | Anthropic 兼容接口 |
| `/v1beta/*` | Gemini 兼容接口 |
| `/api/v1/admin/*` | 管理 API |
| `/api/v1/settings/*` | 公开和系统设置 |
| Clarkaitoy 内部 API | 设备策略、用量和审计同步 |

内部 API 必须使用服务身份鉴权、请求签名和网络访问控制。

## 9. 测试

```powershell
Set-Location backend
go test -tags=unit ./...
go test -tags=integration ./...
golangci-lint run ./...
```

需要覆盖：

- 账号选择、路由、故障转移和熔断。
- 配额、余额、预留和计费边界。
- 模型协议转换和流式响应。
- Clarkaitoy 策略、审计和错误映射。
- 上游 rebase 后的兼容测试。

## 10. 上游同步

```powershell
git fetch upstream
git switch main
git rebase upstream/main
git push origin main
```

如冲突集中在 Clarkaitoy 扩展，优先把扩展封装到独立目录或中间件，避免散改上游
核心代码。
