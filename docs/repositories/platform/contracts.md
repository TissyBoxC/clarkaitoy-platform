# 共享契约代码文档

## 1. 定位

`packages/contracts` 是所有仓库共享的版本化契约目录，负责跨端数据结构和协议
版本，不放具体业务实现、密钥、环境地址或供应商配置。

## 2. 目录结构

```text
packages/contracts/
  openapi/             HTTP API Schema
  events/              异步事件 Schema
  mqtt/                MQTT Topic 和 Payload
  capabilities/        设备能力标识
  generated/           Dart、TypeScript、Go 生成代码
  schemas/             JSON Schema、AsyncAPI 和兼容性配置
```

## 3. 契约范围

| 契约 | 内容 |
| --- | --- |
| HTTP | 家长端、管理端、设备 API 的请求、响应和错误 |
| Event | 设备、内容、OTA、审计、用量和异步任务事件 |
| MQTT | Topic、QoS、Payload、幂等键和保留消息 |
| Capability | 音频、WiFi、摄像头、屏幕、触摸、4G、电池、动作、定位等 |
| Error | 统一错误码、错误类型和重试语义 |
| Version | `schema_version`、兼容策略和废弃周期 |

## 4. 能力标识

当前基础能力包括：

```json
{
  "schema_version": "1.0.0",
  "capabilities": [
    "audio_input",
    "audio_output",
    "wifi",
    "camera",
    "display",
    "touch",
    "led",
    "battery",
    "cellular_4g",
    "motion"
  ]
}
```

后续必须补充：

```text
bluetooth_audio
video_call
location
geofence
sos
multi_device
```

## 5. 版本规则

- 每个契约包含 `schema_version`。
- 破坏性变更必须提升主版本，并提供迁移说明。
- 新增可选字段提升次版本，消费者必须忽略未知字段。
- 删除字段前必须先经过废弃期和兼容性检查。
- 不允许不同仓库自行复制同名契约。

## 6. 生成代码

| 目标 | 生成方式 |
| --- | --- |
| Dart | `openapi-generator` 或 `json_serializable` |
| TypeScript | `openapi-typescript` 或 `orval` |
| Go | `oapi-codegen` 或 `go-jsonschema` |
| JSON Schema | `quicktype`、`jsonschema` 或项目脚本 |

生成代码只读，不允许手工修改。生成前必须通过 Schema 校验、兼容性检查和代码
格式检查。

## 7. 安全约束

- 不放密钥、令牌、连接串和真实环境 URL。
- 不放儿童真实数据、设备证书和内部供应商账号。
- 错误响应不得暴露堆栈、SQL 或供应商原始错误。
- MQTT Topic 和设备标识必须支持脱敏和追踪。
- 事件必须定义生产者、消费者、幂等键和重试策略。

## 8. 测试

| 测试 | 内容 |
| --- | --- |
| Schema 校验 | 契约语法、必填字段和引用完整 |
| 兼容性测试 | 新旧版本消费者和生产者 |
| 生成测试 | Dart、TypeScript、Go 代码可编译 |
| 契约测试 | 服务响应符合 OpenAPI |
| 事件测试 | 幂等、重试、顺序和死信 |

## 9. 删除方法

1. 确认没有消费者和生成代码。
2. 检查跨仓库依赖和版本记录。
3. 移除 Schema、生成配置和生成代码。
4. 运行兼容性、生成和服务契约测试。
