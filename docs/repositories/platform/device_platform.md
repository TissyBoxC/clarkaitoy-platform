# 设备与家庭业务服务代码文档

## 1. 定位

`services/device_platform` 是 Clarkaitoy 的中心业务服务，负责家长账号、家庭、
儿童、设备生命周期、家长策略、内容、OTA、遥测、审计和通知。

它不处理实时音频，不持有模型供应商密钥，不把 `sub2api` 的业务数据复制到儿童
业务表中。

## 2. 目标目录

```text
services/device_platform/
  cmd/
    device-platform/
  internal/
    app/
    config/
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
      notification/
      companion/
      learning/
      vision/
      display/
      input/
      connectivity/
      power/
      motion/
      wearable/
      device_mesh/
    platform/
      database/
      cache/
      messaging/
      observability/
      security/
      storage/
    transport/
      http/
      mqtt/
      ws/
    contracts/
  migrations/
  configs/
  test/
    integration/
```

每个业务模块统一采用：

```text
modules/<module_name>/
  domain/          实体、值对象、错误、接口
  service/         用例和业务规则
  handler/         HTTP、MQTT、事件入口
  repository/      数据访问接口和实现
  module.go        依赖装配和生命周期
```

## 3. 技术选型

完整版本和理由见
[第三方库选型](../technology-selection.md#1-go)。

| 场景 | 默认选择 | 说明 |
| --- | --- | --- |
| 语言 | Go 1.27.1 | 当前服务基线 |
| HTTP | Gin | 生态和中间件最成熟 |
| ORM | GORM | schema、关联和迁移 |
| 复杂查询 | sqlc + pgx | 类型安全 SQL |
| 数据库 | PostgreSQL 16 | 事务和关系数据 |
| 缓存 | Redis 7 | 会话、缓存、限流 |
| 消息 | NATS JetStream | 异步任务和事件 |
| MQTT | Paho MQTT | 设备通信 |
| 配置 | Viper | 多格式和环境变量 |
| 日志 | slog + zerolog | 标准接口和高性能 JSON |
| 迁移 | Goose | schema 版本演进 |
| JWT | golang-jwt | 家长、管理和设备令牌 |

## 4. 模块职责

### 4.1 P0

| 模块 | 职责 |
| --- | --- |
| `auth` | 家长、管理员、设备身份、令牌、MFA、会话和审计 |
| `family` | 家庭、成员、邀请、角色和关系 |
| `child` | 儿童档案、年龄、兴趣、内容等级和隐私字段 |
| `device` | 注册、激活、绑定、解绑、能力协商、证书和状态 |
| `parent_policy` | 时长、禁用时段、内容、音量和单次时长 |
| `content` | 内容包、音频、文本、图片、分龄、审核、发布和下载 |
| `ai_gateway` | 语音网关配置、模型策略、调用标签和供应商状态适配 |
| `ota` | 固件包、灰度、设备组、任务、回滚和失败统计 |
| `telemetry` | 心跳、指标、日志、网络质量、温度和使用报告 |
| `audit` | 管理操作、登录、敏感数据访问和变更记录 |
| `notification` | 推送、短信、邮件、站内消息和远程留言 |

### 4.2 P1

| 模块 | 职责 |
| --- | --- |
| `companion` | 情绪回应、鼓励、成长记录、习惯提醒和长期记忆授权 |
| `learning` | 英语、课程、单词、句型和发音反馈数据 |
| `vision` | 摄像头能力、图片任务、图片隐私和视觉结果索引 |
| `diagnostics` | 设备日志、崩溃、网络和温度诊断聚合 |

### 4.3 P2

| 模块 | 职责 |
| --- | --- |
| `display` | 屏幕、亮度、色温、动画和视频元数据 |
| `input` | 触摸、手势和触觉反馈配置 |
| `connectivity` | eSIM、流量、漫游、远程唤醒和网络切换 |
| `power` | 电池、充电、低功耗和休眠策略 |
| `motion` | 电机、舵机、动作编排和安全限制 |
| `wearable` | 定位、电话、电子围栏和 SOS |
| `device_mesh` | 家庭联动、设备组网和内容同步 |

## 5. 分层规则

- `domain` 不依赖 Gin、GORM、MQTT 或任何数据库实现。
- `service` 只依赖 domain 接口，不直接访问 HTTP 请求或数据库连接。
- `handler` 负责协议转换、输入校验和错误映射。
- `repository` 负责持久化，禁止被其他模块直接引用。
- 模块之间通过公开 service 接口或事件交互。
- 跨模块读取数据必须通过拥有方接口，不允许直接连表查询。
- 模块删除后，应用组合根仍必须能启动。

## 6. 数据与迁移

建议 schema 分组：

```text
identity_*       账号、令牌、会话、角色
family_*         家庭、成员、邀请
child_*          儿童档案、兴趣、隐私授权
device_*         设备、绑定、能力、证书、状态
policy_*         家长策略、时段和内容等级
content_*        内容、主题包、审核和发布
ai_gateway_*     模型策略、调用标签和供应商状态
ota_*            固件、灰度、任务和回滚
telemetry_*      心跳、指标、日志和质量数据
audit_*          操作、访问和变更记录
notification_*   通道、消息、模板和投递结果
```

每个模块拥有自己的迁移文件和 repository。删除模块时，只删除对应表或保留只读
归档，不影响其他模块。

## 7. API 与协议

### 7.1 HTTP

```text
/api/parent/v1/*
/api/admin/v1/*
/api/device/v1/*
/api/internal/v1/*
```

所有请求必须携带：

```text
Authorization
X-Request-ID
Traceparent
```

响应统一包含：

```json
{
  "schema_version": "1.0.0",
  "request_id": "req-001",
  "data": {},
  "error": null
}
```

### 7.2 MQTT/TLS

设备 Topic 由共享契约定义，平台端负责：

- 心跳、在线状态和能力上报。
- 家长策略、内容和远程消息下发。
- OTA 任务和进度回传。
- 遥测、诊断和错误上报。
- 幂等消费、重试、死信和告警。

### 7.3 WebSocket

平台 WebSocket 只用于设备状态和异步任务进度，实时语音使用
`services/voice_gateway`。

## 8. AI 网关适配

`ai_gateway` 模块不直接实现模型调用，它负责：

- 为 `voice_gateway` 提供儿童安全策略和模型白名单。
- 记录设备、儿童、模型、用途和策略版本的请求标签。
- 管理供应商状态、降级策略、配额映射和内部管理 API。
- 通过 `sub2api_fork` 获取模型、账号池、额度和调用统计。
- 不向 `sub2api` 写入儿童姓名、生日、头像、家庭关系或设备密钥。

## 9. 安全与隐私

- 公网通信使用 TLS，设备使用独立证书或设备令牌。
- 儿童敏感字段最小化采集，日志统一脱敏。
- 文件、图片和内容下载使用短时签名 URL。
- 管理操作、数据导出、删除、配额和权限变更必须审计。
- 所有外部输入使用结构化校验，不允许 SQL 拼接。
- 限流、熔断和幂等由平台中间件统一实现。

## 10. 测试

| 层级 | 内容 |
| --- | --- |
| 单元测试 | domain、service、校验、错误映射 |
| repository 测试 | 使用临时 PostgreSQL 验证查询和迁移 |
| HTTP 测试 | 路由、鉴权、输入校验和错误格式 |
| MQTT 测试 | Topic、Payload、幂等、重试和死信 |
| 集成测试 | 家庭、设备、策略、OTA、内容闭环 |
| 契约测试 | OpenAPI、MQTT 和事件 Schema |

## 11. 删除方法

1. 从模块注册表移除模块。
2. 移除路由、MQTT Topic 订阅、事件消费者和定时任务。
3. 移除模块迁移文件或迁移到归档策略。
4. 移除专属依赖和配置。
5. 运行单元、集成和契约测试，验证其他模块仍可启动。
