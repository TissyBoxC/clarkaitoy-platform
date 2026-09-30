# 后台服务端开发文档

## 1. 项目定位

后台服务端是 Clarkaitoy 的中心业务平台，为家长 App、Web 管理端和游戏机本体提供 API、设备通信、AI 网关、内容、OTA 和安全能力。

## 2. 在项目中的组成

```text
services/
  platform_api/
    cmd/server/
    internal/
      modules/
        auth/
        family/
        child/
        device/
        parent_policy/
        content/
        ai_gateway/
        ota/
        telemetry/
        audit/
      platform/
        database/
        cache/
        messaging/
        observability/
        security/
    migrations/
    configs/

  sub2api_fork/
    backend/
    frontend/
    deploy/
```

首发建议使用模块化单体，而不是立即拆微服务。每个业务模块有独立接口和目录，但共享同一个部署单元。达到设备规模或团队规模后，再拆分 `ai_gateway`、`ota_service` 和 `telemetry_service`。

## 3. 技术栈

| 层级 | 推荐技术 |
| --- | --- |
| 后端语言 | Go 1.27 |
| HTTP 框架 | Gin 或 Echo |
| 数据访问 | Ent 或 sqlc |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 消息 | NATS JetStream 或 Redis Streams |
| 设备通信 | MQTT/TLS + WebSocket |
| MQTT Broker | EMQX |
| 对象存储 | S3、MinIO 或兼容服务 |
| 身份认证 | JWT + Refresh Token |
| API 文档 | OpenAPI 3 |
| 观测 | OpenTelemetry + Prometheus + Grafana |
| 日志 | Zap + Loki 或 ELK |
| 追踪 | Tempo 或 Jaeger |
| 部署 | Docker Compose，后续 Kubernetes |

## 4. 推荐实用工具包

### Go 基础

| 包名 | 用途 |
| --- | --- |
| `github.com/gin-gonic/gin` | HTTP 服务 |
| `entgo.io/ent` | ORM 和 schema |
| `github.com/google/wire` | 依赖注入 |
| `github.com/spf13/viper` | 配置 |
| `github.com/redis/go-redis/v9` | Redis |
| `github.com/lib/pq` | PostgreSQL |
| `github.com/golang-jwt/jwt/v5` | JWT |
| `github.com/google/uuid` | ID 生成 |
| `go.uber.org/zap` | 日志 |
| `github.com/robfig/cron/v3` | 定时任务 |
| `github.com/shopspring/decimal` | 金额和额度 |
| `github.com/stretchr/testify` | 单元测试 |

### 设备与实时通信

| 包或服务 | 用途 |
| --- | --- |
| `github.com/eclipse/paho.mqtt.golang` | MQTT 客户端 |
| `github.com/gorilla/websocket` | WebSocket |
| EMQX | MQTT Broker |
| `github.com/coder/websocket` | WebSocket 备用方案 |

### 运维

| 工具 | 用途 |
| --- | --- |
| OpenTelemetry | 指标、日志、追踪 |
| Prometheus | 指标采集 |
| Grafana | 可视化 |
| Loki | 日志聚合 |
| Jaeger 或 Tempo | 分布式追踪 |
| OpenAPI | API 契约 |
| AsyncAPI | 设备消息契约 |

## 5. 功能模块

| 模块 | 负责功能 |
| --- | --- |
| auth | 家长登录、后台登录、令牌、MFA、SSO 预留 |
| family | 家庭、成员、角色、邀请 |
| child | 儿童档案、年龄、内容等级、隐私字段 |
| device | 设备注册、激活、绑定、能力、状态、证书 |
| parent_policy | 时长、时段、内容、音量和远程设置 |
| content | 内容包、主题、素材、审核、发布和下载 |
| ai_gateway | ASR、LLM、TTS、视觉供应商适配 |
| ota | 固件包、灰度、升级任务、回滚和统计 |
| telemetry | 心跳、指标、日志、告警 |
| audit | 管理操作、登录、敏感数据访问追踪 |
| notification | 推送、短信、邮件和站内消息 |

## 6. sub2api 的使用建议

结论：**魔改 fork `sub2api`，但不要把所有 Clarkaitoy 业务写进它。**

推荐采用“双层服务”：

```text
Clarkaitoy platform_api
  ├─ 儿童、家庭、设备、策略、内容、OTA、审计
  └─ ai_gateway 适配层
          │
          ▼
sub2api fork
  ├─ Claude / OpenAI / Gemini / Grok 等上游账号
  ├─ 模型路由、账号池、配额、限流、计费和统计
  └─ OpenAI/Claude/Gemini 兼容接口
```

理由：

- `sub2api` 已经解决多供应商接入、账号池、配额、路由和用量统计。
- 直接复用可以明显减少 AI 基础设施开发量。
- 儿童档案、设备生命周期和家长策略与 AI 网关领域不同，不应写进上游核心。
- `sub2api` 是 LGPL-3.0，需要保留许可证、版权声明和对应修改源码。
- 上游更新频繁，应通过 fork + rebase 维护，不建议直接散改上游目录。

推荐魔改范围：

- 增加 Clarkaitoy 专用组织、项目、设备或租户维度。
- 增加适合儿童设备的模型白名单和内容安全策略。
- 增加调用来源、请求标签和审计字段。
- 增加静默降级、超时和模型熔断策略。
- 增加面向 `platform_api` 的内部管理 API。
- 统一调用日志、用量和错误格式。

不建议放在 `sub2api` 的内容：

- 儿童真实姓名、生日、头像和家庭关系。
- 设备配网、MQTT、OTA 和 GPIO。
- 家长使用时长策略和家长端业务逻辑。
- 内容运营、素材审核和课程体系。

## 7. 接口和协议

### 家长端和后台端

```text
HTTPS + JSON
OpenAPI 3
Bearer Token
schema_version
request_id
trace_id
```

### 游戏机本体

```text
MQTT/TLS：状态、策略、OTA、心跳
WebSocket：流式语音和实时会话
HTTPS：内容下载、固件下载、鉴权
```

### 设备能力

设备连接时上报 capability：

```json
{
  "device_id": "device-001",
  "firmware_version": "1.0.0",
  "capabilities": [
    "audio_input",
    "audio_output",
    "wifi",
    "camera",
    "display",
    "battery"
  ]
}
```

后台根据 capability 决定返回哪些策略、内容和功能开关。

## 8. 安全与儿童隐私

- 所有公网通信必须使用 TLS。
- 设备使用独立证书或设备令牌，不共享家长账号令牌。
- 不记录儿童原始语音，除明确授权和短期处理需要。
- 对照片、语音、位置等敏感数据做最小化采集。
- 日志脱敏，禁止写入账号密码、设备密钥和完整音频。
- 模型请求必须经过内容安全策略。
- 管理操作必须审计。
- 提供家庭数据导出和删除流程。

## 9. 首发范围

第一版实现：

1. 家长和后台账号体系。
2. 设备和家庭绑定。
3. 儿童档案和基础策略。
4. 设备心跳和状态。
5. AI 网关适配 `sub2api`。
6. 基础内容管理和下载。
7. OTA 发布和设备升级。
8. 审计和基础监控。

## 10. 开发规范

- 模块之间只通过公开 service 或接口调用。
- 禁止跨模块直接读取对方数据表。
- 所有外部输入必须校验。
- 所有接口必须带 `request_id` 和 `schema_version`。
- 所有异步任务必须可重试和幂等。
- 所有 AI 调用必须记录成本、延迟、模型、错误和来源。
- 所有设备消息必须可追踪到设备、固件和策略版本。
- 删除某个业务模块时，不能破坏其他模块的编译和启动。
