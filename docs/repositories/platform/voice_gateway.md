# 实时语音服务代码文档

## 1. 定位

`services/voice_gateway` 负责设备实时音频会话、ASR、LLM、TTS、内容安全、会话
状态和用量统计。它是设备与模型供应商之间的隔离层，不保存家庭和设备业务数据。

## 2. 目标目录

```text
services/voice_gateway/
  cmd/
    voice-gateway/
  internal/
    app/
    config/
    transport/
      http/
      websocket/
      video/
    session/
      manager/
      state_machine/
      context/
    audio/
      frame/
      codec/
      buffer/
      playback/
      vad/
      aec/
    adapter/
      asr/
        aliyun/
        tencent/
        volcano/
        openai/
        local_whisper/
      tts/
        aliyun/
        tencent/
        volcano/
        openai/
        local_piper/
      realtime/
        openai_realtime/
      vision/
    llm/
      sub2api_client/
    security/
      content_policy/
      moderation/
      prompt_guard/
      image_privacy/
    telemetry/
    usage/
    platform/
      cache/
      clock/
      messaging/
      observability/
    contracts/
      websocket/
      internal-api/
  test/
    integration/
    load/
```

## 3. 技术选型

完整版本和理由见
[第三方库选型](../technology-selection.md#1-go)。

| 场景 | 默认选择 | 说明 |
| --- | --- | --- |
| HTTP | Gin | 健康检查、内部 API 和管理接口 |
| WebSocket | Coder WebSocket 或 Gorilla WebSocket | 选择一个实现，不在同一服务混用 |
| 音频编解码 | Opus | 低带宽语音 |
| 回声和降噪 | SpeexDSP | AEC、NS、AGC |
| ASR | 云供应商或本地 Whisper | 通过 adapter 隔离 |
| LLM | `sub2api_fork` | 模型统一入口 |
| TTS | 云供应商或 Piper | 按语言和设备能力选择 |
| 缓存 | Redis 7 | 会话状态、限流、并发控制 |
| 消息 | NATS JetStream | 用量、事件和异步任务 |
| 配置 | Viper | 服务配置和供应商配置 |
| 日志 | slog + zerolog | 结构化日志 |
| 指标 | OpenTelemetry + Prometheus | 延迟、并发、错误和用量 |

## 4. 功能模块

### 4.1 P0 实时对话

| 模块 | 职责 |
| --- | --- |
| `websocket` | 鉴权、连接、心跳、帧解析、背压和断开 |
| `session` | 会话创建、状态机、超时、取消和并发控制 |
| `audio/frame` | 音频帧格式、序号、时间戳和校验 |
| `audio/codec` | Opus 编解码和采样率转换 |
| `audio/buffer` | 环形缓冲、抖动控制和背压 |
| `audio/playback` | 播放队列、优先级、打断和恢复 |
| `asr` | 语音识别供应商适配 |
| `llm/sub2api_client` | 多轮对话、流式输出、重试和降级 |
| `tts` | 语音合成、音色、语速和播放 |
| `security` | 输入和输出内容安全、Prompt Guard |
| `usage` | 模型、Token、时长、延迟和错误统计 |

### 4.2 P1

| 模块 | 职责 |
| --- | --- |
| `vad` | 语音活动检测和自动断句 |
| `aec` | 回声消除、降噪和远场增强 |
| `continuous_conversation` | 唤醒后连续会话和超时 |
| `companion` | 情绪回应、鼓励和长期记忆上下文 |
| `learning` | 英语口语练习、发音反馈和课程上下文 |
| `vision` | 图片任务、识物、绘本和拍照问答 |
| `diagnostics` | 会话质量、网络质量和供应商延迟 |

### 4.3 P2

| 模块 | 职责 |
| --- | --- |
| `video` | 视频通话、媒体协商、音视频同步和远程陪伴 |
| `display_events` | 屏幕动画、点读和视频控制事件 |
| `motion_events` | 动作、跳舞和停止指令 |
| `device_mesh_events` | 家庭联动和多设备会话路由 |

## 5. 会话状态机

```text
idle
  -> connecting
  -> authenticating
  -> listening
  -> recognizing
  -> thinking
  -> synthesizing
  -> playing
  -> listening
  -> closing
  -> closed
```

任何阶段都允许进入：

```text
cancelled
timed_out
failed
```

约束：

- 状态转移必须由状态机统一处理。
- 音频帧必须有单调序号和时间戳。
- 取消后不得继续向设备发送旧响应。
- 供应商失败必须降级或明确结束，不允许静默挂起。
- 会话并发、单设备连接数和总连接数必须可配置。

## 6. 音频协议

WebSocket 帧建议包含：

```json
{
  "schema_version": "1.0.0",
  "type": "audio_frame",
  "session_id": "session-001",
  "sequence": 12,
  "timestamp_ms": 1735689600000,
  "codec": "opus",
  "sample_rate": 16000,
  "channels": 1,
  "payload_base64": "..."
}
```

协议规则：

- 首帧必须携带设备身份、固件版本、能力和会话参数。
- 控制帧和音频帧分离，控制帧不得阻塞音频处理。
- 服务端必须限制帧大小、速率、并发和会话时长。
- 所有错误返回稳定错误码，不把供应商错误直接暴露给设备。

## 7. 安全与儿童保护

- 设备使用短期会话令牌，不共享家长账号令牌。
- 输入 ASR 文本、图片元数据和输出 TTS 文本都必须经过内容策略。
- 原始音频默认不落盘；需要短期处理时使用加密临时存储并自动删除。
- 图片必须遵守 `privacy_guard` 和 `image_privacy` 授权。
- Prompt Guard 防止越权提示、供应商密钥泄露和不适龄输出。
- 供应商请求、响应、模型、延迟、Token 和错误必须记录关联 ID。
- 日志禁止记录完整音频、图片、儿童真实姓名和家庭关系。

## 8. sub2api 集成

`llm/sub2api_client` 负责：

- 调用 `sub2api` 的 OpenAI、Anthropic 或 Gemini 兼容接口。
- 传递 Clarkaitoy 租户、设备、用途、策略和请求标签。
- 处理流式响应、超时、重试、熔断和降级。
- 记录模型、Token、成本、延迟和错误。
- 在 `sub2api` 不可用时切换到允许的备用模型或明确失败。

不允许在语音网关中保存模型供应商账号和密钥。供应商凭据由 `sub2api_fork`
管理。

## 9. 观测

核心指标：

```text
voice_session_total
voice_session_active
voice_session_duration_seconds
voice_asr_latency_seconds
voice_llm_first_token_seconds
voice_tts_first_byte_seconds
voice_playback_underrun_total
voice_provider_error_total
voice_content_block_total
voice_usage_tokens_total
voice_usage_audio_seconds
```

## 10. 测试

| 层级 | 内容 |
| --- | --- |
| 单元测试 | 状态机、帧解析、缓冲、队列、安全策略 |
| 供应商契约测试 | ASR、TTS、LLM 适配器接口 |
| WebSocket 测试 | 鉴权、心跳、背压、取消和重连 |
| 集成测试 | 音频到文本到模型到 TTS 的闭环 |
| 压测 | 并发会话、CPU、内存、延迟和丢帧 |
| 故障测试 | 供应商超时、限流、断网和熔断 |

## 11. 删除方法

1. 从服务组合根移除模块初始化和路由。
2. 移除对应供应商配置和依赖。
3. 移除会话、安全和用量中的可选钩子。
4. 运行单元、集成和压测，验证基础对话链路仍可用。
