# 第三方库选型

本文件是 Clarkaitoy 各仓库的通用选型基线。实现功能时必须优先复用成熟库，
不得自行编写已有成熟解决方案的 Web 框架、ORM、状态管理、HTTP 客户端、UI 组件、
消息协议、硬件驱动或通用工具。

版本号是选型建议，实际落地时以项目锁文件和兼容性验证为准。引入依赖前必须检查
许可证、维护状态、发布频率、目标平台支持和对固件体积的影响。

## 1. Go

### 1.1 运行服务

| 场景 | 推荐库 | 导入路径 | 定位与选用理由 | 适用场景 |
| --- | --- | --- | --- | --- |
| Web 框架 | Gin | `github.com/gin-gonic/gin` | Go 生态事实标准，资料和中间件最多 | 平台 API、设备网关、后台接口 |
| Web 框架 | Echo | `github.com/labstack/echo/v4` | API 简洁，生命周期和错误处理清晰 | 偏好轻量路由和显式中间件 |
| Web 框架 | Fiber | `github.com/gofiber/fiber/v3` | 基于 Fasthttp，性能激进，API 接近 Express | 高吞吐内部服务，不依赖 net/http 中间件 |
| ORM/数据库 | GORM | `gorm.io/gorm` | 生态成熟，迁移、关联、钩子和 PostgreSQL 支持完整 | 业务实体和 repository |
| SQL 生成 | sqlc | `github.com/sqlc-dev/sqlc` | SQL 可审查，生成类型安全 Go 代码 | 查询复杂和高性能读写路径 |
| PostgreSQL 驱动 | pgx | `github.com/jackc/pgx/v5` | PostgreSQL 原生能力、连接池和批量操作完善 | PostgreSQL、GORM/sqlc 底层驱动 |
| Redis | go-redis | `github.com/redis/go-redis/v9` | Redis 官方推荐客户端，支持集群和 Lua | 会话、限流、缓存、分布式锁 |
| 消息 | NATS Go | `github.com/nats-io/nats.go` | JetStream 提供持久化和消费确认 | 异步任务、事件流、设备遥测 |
| MQTT | Paho MQTT | `github.com/eclipse/paho.mqtt.golang` | 经生产验证的 MQTT 客户端 | 设备状态、策略和 OTA 下发 |
| WebSocket | Gorilla WebSocket | `github.com/gorilla/websocket` | 成熟稳定，应用面广 | 实时控制、语音元数据 |
| WebSocket | Coder WebSocket | `github.com/coder/websocket` | 上下文友好，API 简洁 | 新服务和无 Gorilla 兼容要求 |
| 配置 | Viper | `github.com/spf13/viper` | 多格式、环境变量和远程配置 | 服务配置、配置文件和环境覆盖 |
| 日志 | slog | `log/slog` | Go 标准库，结构化日志和扩展 Handler 足够 | 新服务优先选择 |
| 日志 | zerolog | `github.com/rs/zerolog` | 零分配 JSON 日志，性能突出 | 高吞吐日志和 JSON 管线 |
| CLI | Cobra | `github.com/spf13/cobra` | kubectl、Hugo 等生产项目使用 | 管理 CLI、迁移工具、运维命令 |
| 测试断言 | Testify | `github.com/stretchr/testify` | assert/require/mock 语义清晰 | 单元测试和集成测试 |
| 模糊测试 | Go fuzz | `testing.F` | 标准库支持，CI 成本低 | 协议解析、输入校验 |
| 泛型工具 | samber/lo | `github.com/samber/lo` | Lodash 风格泛型集合和 Map 工具 | 集合转换、过滤和分组 |
| 错误聚合 | go-multierror | `github.com/hashicorp/go-multierror` | 统一收集并发子任务错误 | 批量初始化、批量任务 |
| 依赖注入 | Wire | `github.com/google/wire` | 编译期生成依赖图，无运行时反射 | 大型服务模块组装 |
| UUID | google/uuid | `github.com/google/uuid` | UUID 和随机标识生成 | API、事件、请求标识 |
| 定时任务 | cron | `github.com/robfig/cron/v3` | 成熟 Cron 解析和调度 | 报表、清理、定时推送 |
| 金额/额度 | decimal | `github.com/shopspring/decimal` | 十进制金额和配额计算 | 计费、用量额度 |
| 可观测性 | OpenTelemetry Go | `go.opentelemetry.io/otel` | Trace、Metric、Log 的统一标准 | 跨服务链路和指标 |
| 指标 | Prometheus Go | `github.com/prometheus/client_golang` | Prometheus 官方客户端 | 服务指标和告警 |
| JWT | golang-jwt | `github.com/golang-jwt/jwt/v5` | JWT 解析、签名和验证成熟 | 家长端、管理端和设备令牌 |
| 密码哈希 | bcrypt | `golang.org/x/crypto/bcrypt` | 标准密码哈希实现 | 账号密码 |
| 数据库迁移 | Goose | `github.com/pressly/goose/v3` | SQL/Go 迁移，CI 和 CLI 友好 | PostgreSQL schema 演进 |

### 1.2 Go 选型建议

- 平台 API 首选 Gin + GORM + PostgreSQL + Redis + Viper + slog/zerolog。
- 复杂查询使用 sqlc + pgx，不要为每个查询手写重复的数据访问层。
- 事件和异步任务首选 NATS JetStream；已经使用 Redis 且规模较小时可用 Redis
  Streams。
- 设备侧和平台侧 MQTT 客户端统一使用 Paho MQTT。
- 标准库已经覆盖 `slog`、`errors.Is/As`、`context`、`testing`、`net/http`
  时优先使用标准库，不额外引入同功能依赖。
- CLI、迁移、运维命令统一使用 Cobra 和 Goose，不自建参数解析和迁移框架。

## 2. Flutter

### 2.1 UI 与设计系统

| 场景 | 包名 | 官方/社区 | 版本约束建议 | 定位与选用理由 | 适用场景 |
| --- | --- | --- | --- | --- | --- |
| UI 组件 | `GetWidget` | 社区 | `^7.0.1` | 1000+ 生产级 Widget，覆盖常见移动端组件 | 快速搭建业务页面 |
| UI 组件 | `shadcn_ui` | 社区 | `^0.28.0` | 现代化、可定制、组件源码可控 | 需要品牌化设计系统 |
| UI 组件 | `forui` | 社区 | `^0.21.0` | 语义化组件和一致的主题系统 | 新项目、快速原型 |
| 动画 | `lottie` | 社区 | `^3.3.1` | 播放 After Effects 动画 | 儿童反馈、空状态、加载动画 |
| SVG | `flutter_svg` | 社区 | `^2.2.0` | SVG 图标和插画渲染 | 资源体积和缩放要求 |
| 图片缓存 | `cached_network_image` | 社区 | `^3.4.1` | 网络图片缓存和占位 | 内容封面、头像 |
| 图表 | `fl_chart` | 社区 | `^1.1.1` | 常用统计图表，API 清晰 | 使用报告和趋势 |
| 日历 | `table_calendar` | 社区 | `^3.2.0` | 日历、区间和日程选择 | 使用时长和禁用时段 |

### 2.2 状态、路由和数据

| 场景 | 包名 | 官方/社区 | 版本约束建议 | 定位与选用理由 | 适用场景 |
| --- | --- | --- | --- | --- | --- |
| 状态管理 | `flutter_riverpod` | 社区热门 | `^3.0.3` | 测试友好、组合能力强、无全局单例约束 | 新项目默认选择 |
| 状态管理 | `flutter_bloc` | 社区热门 | `^9.1.1` | 事件和状态显式，团队约束强 | 大型团队和复杂流程 |
| 状态管理 | `signals_flutter` | 社区 | `^6.2.0` | 细粒度响应式和低学习成本 | 小型模块和局部状态 |
| 路由 | `go_router` | 官方推荐 | `^17.0.0` | 声明式路由、深层链接和守卫 | 全部 Flutter 应用 |
| 网络 | `dio` | 社区热门 | `^5.9.0` | 拦截器、取消、上传、超时和错误处理完整 | REST API 和文件传输 |
| 序列化 | `json_serializable` | 官方生态 | `^6.11.1` | 代码生成、类型安全 | API DTO 和持久化模型 |
| 模型生成 | `freezed` | 社区热门 | `^3.2.3` | 不可变模型、联合类型和 copyWith | 领域模型和状态模型 |
| API 生成 | `retrofit` | 社区 | `^4.9.2` | 基于 Dio 的接口生成 | 大型 API 客户端 |
| 本地设置 | `shared_preferences` | 官方生态 | `^2.5.3` | 轻量键值存储 | 普通用户设置 |
| 安全存储 | `flutter_secure_storage` | 社区热门 | `^10.0.0` | Keychain/Keystore 支持 | 令牌、密钥、敏感配置 |
| 本地数据库 | `drift` | 社区热门 | `^2.30.0` | SQLite 类型安全、迁移和查询能力 | 离线内容、历史和复杂数据 |
| 网络状态 | `connectivity_plus` | 官方生态 | `^7.0.0` | 网络类型和连接状态 | 离线提示和重试 |
| 推送 | `firebase_messaging` | 官方生态 | `^16.2.0` | FCM 跨平台推送 | 设备状态和内容通知 |
| 权限 | `permission_handler` | 社区热门 | `^12.0.1` | 移动权限统一接口 | 相机、通知、存储 |
| 国际化 | `flutter_localizations` + `intl` | 官方推荐 | 随 Flutter SDK，`intl ^0.20.2` | 官方本地化和 ICU 格式化 | 中文和未来多语言 |

### 2.3 测试与质量

| 场景 | 包名 | 官方/社区 | 版本约束建议 | 定位与选用理由 |
| --- | --- | --- | --- | --- |
| Lint | `flutter_lints` | 官方推荐 | `^6.0.0` | 官方推荐规则集 |
| Mock | `mocktail` | 社区热门 | `^1.0.4` | 无需代码生成的 mock |
| 端到端 | `integration_test` | 官方 | Flutter SDK | 官方集成测试 |
| E2E | `patrol` | 社区 | `^4.1.0` | 原生交互和端到端测试 |
| 日志 | `logger` | 社区热门 | `^2.6.2` | 简洁分级日志 |
| 网络调试 | `talker_flutter` | 社区 | `^5.1.5` | 应用内日志和网络面板 |

### 2.4 Flutter 选型建议

- 新项目默认采用 Riverpod + Dio + go_router + json_serializable/freezed。
- 大型团队需要强约束时使用 BLoC，不为同一 feature 混用两套全局状态方案。
- UI 优先使用 GetWidget 或 shadcn_ui，自定义设计系统只封装主题和品牌差异。
- 所有 API 响应必须通过生成模型映射，不在 Widget 中直接解析 JSON。
- 所有安全令牌使用 `flutter_secure_storage`，普通设置才使用
  `shared_preferences`。
- 离线内容和历史记录使用 drift，不为复杂查询手写文件格式。

## 3. Vue 3 + Vue Router

### 3.1 核心

| 场景 | npm 包名 | 定位与选用理由 | 适用场景 |
| --- | --- | --- | --- |
| 路由 | `vue-router` | Vue 官方维护，支持 Composition API、懒加载、守卫和动态路由 | 全部 Vue 应用 |
| 状态管理 | `pinia` | Vue 官方推荐，TypeScript 支持好，取代 Vuex | 全局登录、权限、主题状态 |
| 请求状态 | `@tanstack/vue-query` | 缓存、重试、失效、分页和异步状态统一 | 管理台和服务端状态 |
| HTTP | `axios` | 拦截器、取消、上传下载和错误处理成熟 | API 客户端底层 |
| 工具集 | `@vueuse/core` | Composition API 工具、浏览器 API、表单和动画 | 所有 Vue 项目 |
| 表单 | `vee-validate` | 表单状态和校验语义清晰，支持组合式校验 | 管理台复杂表单 |
| Schema 校验 | `zod` | 前后端共享契约校验和类型推导 | API 输入、表单和配置 |
| 日期 | `dayjs` | 体积小，插件丰富 | 时间格式化、筛选和统计 |
| 图表 | `echarts` | 管理台图表生态成熟 | 运营报表和监控 |
| 国际化 | `vue-i18n` | Vue 官方生态国际化方案 | 中文和未来多语言 |

### 3.2 UI 组件库

| 场景 | npm 包名 | 定位与选用理由 | 适用场景 |
| --- | --- | --- | --- |
| 中后台默认 | `element-plus` | 生态成熟、中文文档完整、管理组件齐全 | 运营控制台、后台管理 |
| 企业规范 | `ant-design-vue` | 企业级设计规范和复杂数据展示成熟 | 遵循 Ant Design 的团队 |
| TypeScript 与定制 | `naive-ui` | 完全 TypeScript 编写，主题和性能好 | 高度定制、TS 优先 |
| 移动端 H5 | `vant` | 移动组件完整、体积和无障碍控制好 | 家长端 H5、移动运营页 |
| 轻量灵活 | `@headlessui/vue` + `tailwindcss` | 无样式组件和原子化 CSS | 品牌化、轻量界面 |
| 图标 | `@element-plus/icons-vue` | Element Plus 官方图标 | 配合 Element Plus |

### 3.3 质量工具

| 场景 | npm 包名 | 定位与选用理由 |
| --- | --- | --- |
| Lint | `eslint` | JavaScript/TypeScript/Vue 规则生态核心 |
| 格式化 | `prettier` | 统一代码和 Markdown 格式 |
| 单元测试 | `vitest` | Vite 原生测试，速度快 |
| 组件测试 | `@vue/test-utils` | Vue 官方组件测试入口 |
| E2E | `@playwright/test` | 浏览器端到端和视觉回归 |
| API Mock | `msw` | 浏览器和 Node 共用 Mock |
| Commit 校验 | `@commitlint/cli` | 约束提交信息格式 |

### 3.4 Vue 选型建议

- 管理端默认采用 Vue Router 4 + Pinia + Element Plus + Axios + VueUse。
- 服务端状态统一交给 Vue Query，避免把 API 缓存复制进 Pinia。
- 中后台复杂表格优先使用 Element Plus 表格，数据量和交互复杂时使用
  `@tanstack/vue-table`。
- 所有表单和接口输入使用 VeeValidate + Zod，不在组件中手写重复校验。
- 移动端页面使用 Vant；需要自定义品牌时使用 Headless UI + Tailwind。

## 4. PlatformIO 与 ESP32-S3

### 4.1 通信、数据与基础设施

| 场景 | PlatformIO 库名或组件 | `lib_deps` 写法 | 适用芯片/框架 | 定位与选用理由 |
| --- | --- | --- | --- | --- |
| JSON | ArduinoJson | `bblanchon/ArduinoJson @ ^7.4.2` | ESP32、ESP-IDF/Arduino | 嵌入式 JSON 事实标准，内存可控 |
| MQTT | PubSubClient | `knolleary/PubSubClient @ ^2.8` | ESP32、Arduino | MQTT 客户端成熟，API 简单 |
| MQTT over TLS | ESP-IDF MQTT | ESP-IDF 组件 `mqtt` | ESP32、ESP-IDF | 原生 TLS、事件循环和证书能力 |
| WebSocket | `esp_websocket_client` | ESP-IDF 组件 | ESP32、ESP-IDF | 官方流式连接和重连支持 |
| HTTP | `esp_http_client` | ESP-IDF 组件 | ESP32、ESP-IDF | OTA、下载和 REST API |
| 网络 | `esp_wifi`、`esp_netif` | ESP-IDF 组件 | ESP32-S3、ESP-IDF | 官方联网和网络接口 |
| 事件 | `esp_event` | ESP-IDF 组件 | ESP32、ESP-IDF | 模块间事件解耦 |
| 配置 | `nvs_flash` | ESP-IDF 组件 | ESP32、ESP-IDF | 非易失配置和身份存储 |
| OTA | `esp_ota_ops` | ESP-IDF 组件 | ESP32、ESP-IDF | 官方 OTA 安装和回滚 |
| 定时 | `esp_timer` | ESP-IDF 组件 | ESP32、ESP-IDF | 高精度定时和周期任务 |
| 日志 | `esp_log` | ESP-IDF 组件 | ESP32、ESP-IDF | 官方日志和级别控制 |

### 4.2 音频、语音与显示

| 场景 | PlatformIO 库名或组件 | `lib_deps` 写法 | 适用芯片/框架 | 定位与选用理由 |
| --- | --- | --- | --- | --- |
| 音频管线 | ESP-ADF | ESP-IDF 组件 `esp-adf-libs` | ESP32-S3、ESP-IDF | 录音、播放、编解码和 pipeline |
| 唤醒词 | ESP-SR | ESP-IDF 组件 `esp-sr` | ESP32-S3、ESP-IDF | 官方 WakeNet 和多语种模型 |
| 音频设备 | `esp_codec_dev` | ESP-IDF 组件 | ESP32-S3、ESP-IDF | 编解码器统一抽象 |
| 语音编解码 | Opus | ESP-IDF 组件或 `pschatzmann/arduino-audio-tools` | ESP32-S3、ESP-IDF | 低带宽实时语音 |
| 回声消除 | SpeexDSP | ESP-IDF 组件 `esp-sr` 或独立 port | ESP32-S3、ESP-IDF | AEC、降噪和增益 |
| TFT 显示 | LovyanGFX | `lovyan03/LovyanGFX @ ^1.2.7` | ESP32、Arduino/ESP-IDF | 驱动广泛，性能和兼容性好 |
| TFT 显示 | TFT_eSPI | `bodmer/TFT_eSPI @ ^2.5.43` | ESP32、Arduino | 经典 TFT 驱动，配置丰富 |
| OLED 显示 | Adafruit SSD1306 | `adafruit/Adafruit SSD1306 @ ^2.5.15` | ESP32、Arduino | 经典 OLED 驱动 |
| 显示基础 | Adafruit GFX | `adafruit/Adafruit GFX Library @ ^1.11.11` | ESP32、Arduino | 图形和字体基础库 |
| LVGL | LVGL | `lvgl/lvgl @ ^9.3.0` | ESP32-S3、LVGL | 嵌入式 GUI 事实标准 |
| LVGL 接入 | `esp_lvgl_port` | ESP-IDF 组件 | ESP32-S3、ESP-IDF | 官方 LVGL 显示和触摸接入 |

### 4.3 传感器、输入与执行器

| 场景 | PlatformIO 库名或组件 | `lib_deps` 写法 | 适用芯片/框架 | 定位与选用理由 |
| --- | --- | --- | --- | --- |
| BME280 | Adafruit BME280 | `adafruit/Adafruit BME280 Library @ ^2.3.0` | ESP32、Arduino | 温湿度和气压驱动成熟 |
| Unified Sensor | Adafruit Unified Sensor | `adafruit/Adafruit Unified Sensor @ ^1.1.15` | ESP32、Arduino | Adafruit 传感器统一接口 |
| DHT | DHT sensor library | `adafruit/DHT sensor library @ ^1.4.6` | ESP32、Arduino | DHT11/DHT22 经典驱动 |
| 按键 | EasyButton | `easybtn/EasyButton @ ^2.0.4` | ESP32、Arduino | 消抖、单击、双击和长按 |
| 触摸 | `esp_lcd_touch` | ESP-IDF 组件 | ESP32-S3、ESP-IDF | 官方触摸屏适配框架 |
| 舵机 | ESP32 Servo | `madhephaestus/ESP32Servo @ ^3.0.6` | ESP32、Arduino | 舵机 PWM 控制 |
| 步进电机 | AccelStepper | `waspinator/AccelStepper @ ^1.64` | ESP32、Arduino | 步进电机加减速和运动控制 |
| 摄像头 | `esp_camera` | ESP-IDF 组件 | ESP32-S3、ESP-IDF | 官方摄像头驱动 |
| 4G | TinyGSM | `vshymanskyy/TinyGSM @ ^0.12.0` | ESP32、Arduino | 多模组蜂窝网络抽象 |
| 定位 | TinyGPSPlus | `mikalhart/TinyGPSPlus @ ^1.1.0` | ESP32、Arduino | NMEA 解析和定位数据 |

### 4.4 测试与调试

| 场景 | 工具或库 | 使用方式 | 适用场景 |
| --- | --- | --- | --- |
| 单元测试 | PlatformIO Unit Testing | `pio test -e native` 或目标环境 | 纯逻辑模块和协议解析 |
| 静态分析 | `cppcheck` | CI 和本地检查 | C 代码缺陷 |
| 格式化 | `clang-format` | pre-commit 和 CI | C/C++ 格式 |
| 串口监视 | PlatformIO Monitor | `pio device monitor` | 设备日志和错误诊断 |
| 烧录 | `esptool.py` | `pio run -t upload` | 固件和芯片信息 |
| MQTT 调试 | MQTTX | 桌面或 CLI | Topic、Payload 和 TLS |
| 网络分析 | Wireshark | 抓包和 TLS 分析 | 设备网络问题 |
| 协议校验 | JSON Schema | CI 验证共享契约 | MQTT、事件和 API |

### 4.5 PlatformIO 选型建议

- ESP-IDF 原生能力优先使用 ESP-IDF 官方组件，不引入 Arduino 封装替代。
- MQTT/TLS、OTA、网络、日志和事件使用 ESP-IDF 官方组件。
- 上层业务只依赖模块公开头文件和适配接口，不直接依赖具体驱动库。
- JSON 统一使用 ArduinoJson；协议结构必须先在共享契约中定义。
- 屏幕统一使用 LVGL，显示驱动优先 LovyanGFX，OLED 使用 Adafruit SSD1306。
- 传感器和输入设备优先 Adafruit 系列、EasyButton 等成熟驱动。
- 任何 `lib_deps` 依赖都必须锁定主版本，并在移除对应模块时同步移除依赖。

## 5. 跨技术栈约束

- 不手写已有成熟库覆盖的 Web 框架、ORM、状态管理、表单校验、JSON、MQTT、
  LVGL、传感器驱动和 OTA 逻辑。
- 不推荐无人维护、star 数极低或已经废弃的库。
- 不为了减少依赖而复制第三方库源码。
- 固件模块的 `lib_deps` 必须与模块开关绑定，关闭模块时不能下载或链接无关库。
- 所有第三方依赖必须在 CI 中执行许可证、安全和版本检查。
