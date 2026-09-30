# 项目缺口与下一阶段计划

## 1. 当前基线

当前工作区已经具备可运行的骨架：

- `apps/parent_app`：Flutter 应用可启动，已有 feature-first 目录、路由、主题、配置边界和基础页面。
- `apps/admin_web`：Vue 3 管理端可构建，已有应用壳、路由、状态容器、HTTP 边界和基础页面。
- `services/device_platform`：Go 服务可通过 `go test ./...` 和 `go build ./...`，已有九个业务模块的四层目录。
- `services/voice_gateway`：Go 服务可通过测试和构建，已有会话状态机、音频边界、ASR/TTS/LLM 适配接口。
- `firmware`：独立 PlatformIO 仓库，已配置 ESP32-S3 N16R8、16 MB Flash、8 MB OPI PSRAM 和 `bgen` 构建目录。
- `packages/contracts`：已建立共享契约的版本化目录。

`services/sub2api` 是独立上游仓库，不应在这里直接混入 Clarkaitoy 业务代码。

## 2. 家长端缺口

家长端当前是可运行的导航壳，以下功能仍是占位：

- 登录、注册、验证码、令牌刷新和会话恢复。
- 家庭、成员、邀请和角色权限。
- 儿童档案、兴趣、年龄和内容等级。
- 设备扫码、蓝牙发现、配网、绑定和解绑。
- 使用时长、禁用时段、音量上限和内容策略。
- 使用报告、远程留言、内容下载和 OTA 升级。
- 隐私设置、数据导出和数据删除。

每个 feature 后续应补齐 `application/`、`domain/`、`data/` 和对应测试。当前只建立了跨 feature 的公共边界和两个可运行页面。

## 3. 管理端缺口

管理端当前是可运行的运营控制台壳，以下模块仍是占位：

- 管理员登录、MFA、会话和 RBAC。
- 家庭、儿童、设备和内容管理页面。
- AI 网关的模型、路由、配额和调用日志视图。
- OTA 发布、灰度、回滚和失败统计。
- 审计、监控、告警和系统配置。
- Element Plus、表格、图表和统一表单组件接入。

后续接入接口时应保持 feature 独立，不能让页面直接读取供应商密钥或设备长期密钥。

## 4. 后台服务缺口

`device_platform` 已有模块边界和健康检查，但尚未实现：

- PostgreSQL schema、迁移和 repository 实现。
- Redis 会话、缓存和限流。
- MQTT 连接、TLS、订阅、重连和幂等消费。
- 设备注册、证书、绑定和 capability 协商。
- 家长策略、内容、OTA、遥测、审计和通知的业务 API。
- OpenAPI 路由注册和统一错误响应。
- 集成测试、数据库测试和可观测性指标。

`voice_gateway` 已有接口和状态机，但尚未实现：

- WebSocket 鉴权、音频帧协议、心跳和背压。
- Opus 编解码、环形缓冲、播放队列和 VAD。
- ASR、TTS、实时语音供应商的真实连接。
- `sub2api` 的流式请求、错误降级、重试和用量记录。
- 儿童内容审核、Prompt Guard 和输出安全检查。
- 会话并发控制、限流、指标和高负载测试。

## 5. 固件缺口

固件当前只有最小 ESP-IDF 启动点。后续需要按可删除组件补齐：

- `system_core`、`audio_io`、`voice_wake` 和 `voice_session`。
- ASR、LLM、TTS 网络客户端与重连。
- 内容库、家长策略、MQTT 和 OTA。
- NVS 设备身份、TLS 证书和隐私保护。
- 可选摄像头、屏幕、触摸、LED、电池、4G 和动作模块。

每个可选模块都必须有独立 Kconfig 开关，并在移除后不影响无关模块编译。

## 6. 跨项目缺口

以下内容尚未完成：

- 统一的 OpenAPI、事件、MQTT 和 capability 契约版本。
- Dart、TypeScript 和 Go 共享类型或生成客户端。
- Docker Compose、配置样例、数据库和本地依赖编排。
- CI 中的 Flutter、Go、Vue 和 PlatformIO 质量门禁。
- 密钥管理、设备证书轮换、隐私合规和审计方案。
- `services/sub2api` 的 fork 命名、许可证保留、升级 rebase 策略和业务适配层。

## 7. 推荐实施顺序

1. 固化身份、家庭、儿童、设备和 capability 契约。
2. 完成 `device_platform` 的数据库、认证和设备绑定闭环。
3. 完成 `voice_gateway` 的 WebSocket、ASR、`sub2api` 和 TTS 闭环。
4. 接入家长端登录、绑定、儿童档案和策略。
5. 接入管理端的设备、内容、OTA 和审计页面。
6. 在固件中按开关逐项启用音频、网络、策略和 OTA。
7. 最后补充监控、压测、隐私合规、灰度发布和灾难恢复。
