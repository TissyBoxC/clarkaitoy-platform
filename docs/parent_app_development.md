# 家长控制端开发文档

## 1. 项目定位

家长控制端是 Clarkaitoy 面向家长和监护人的移动应用，使用 Flutter 构建。

它负责以下核心事情：

- 登录和家庭账号管理
- 绑定和管理游戏机本体
- 创建和管理儿童档案
- 配置内容权限和使用时长
- 查看使用报告和设备状态
- 发送远程留言和远程内容
- 管理固件升级
- 配置隐私、安全和儿童保护选项

家长控制端不负责实时 AI 对话、不直接连接游戏机本体，也不保存设备密钥。

## 2. 在项目中的组成

```text
apps/
  parent_app/
    lib/
      app/                 应用启动、主题、路由
      core/                网络、存储、日志、错误处理
      features/
        auth/
        family/
        child_profile/
        device/
        parent_policy/
        usage_report/
        remote_message/
        content_management/
        ota/
        settings/
      shared/              通用组件和设计系统
      l10n/                国际化
    test/
```

推荐按 feature-first 组织，不按 controller、service、model 这种全局技术层组织。每个 feature 独立拥有页面、状态、API 和测试，方便删除功能。

## 3. 技术栈

| 层级 | 推荐技术 |
| --- | --- |
| 开发框架 | Flutter 3.x |
| 语言 | Dart 3.x |
| 状态管理 | Riverpod 或 Bloc |
| 路由 | go_router |
| 网络请求 | dio |
| 数据模型 | freezed + json_serializable |
| 本地存储 | shared_preferences + flutter_secure_storage |
| 依赖注入 | get_it 或 Riverpod Provider |
| 本地数据库 | drift 或 isar |
| 国际化 | flutter_localizations + intl |
| 日志 | logger |
| 测试 | flutter_test + integration_test + mocktail |

推荐组合：

- 小团队快速开发：Riverpod + go_router + dio + freezed。
- 大型团队强约束：Bloc + go_router + dio + freezed。

## 4. 推荐实用工具包

### UI 和交互

| 包名 | 用途 |
| --- | --- |
| `flutter_screenutil` | 多屏尺寸适配 |
| `flutter_svg` | SVG 图标和插画 |
| `cached_network_image` | 图片缓存 |
| `lottie` | 轻量动画 |
| `shimmer` | 加载占位效果 |
| `flutter_form_builder` | 表单和校验 |
| `pinput` | 验证码输入 |
| `carousel_slider` | 儿童档案等内容轮播 |
| `fl_chart` | 使用时长和报告图表 |
| `table_calendar` | 日历和时间计划 |
| `flutter_local_notifications` | 本地提醒 |

### 数据和基础设施

| 包名 | 用途 |
| --- | --- |
| `dio` | HTTP API |
| `retrofit` | API 代码生成 |
| `connectivity_plus` | 网络状态 |
| `firebase_messaging` | 推送消息 |
| `flutter_secure_storage` | 令牌和敏感信息 |
| `shared_preferences` | 普通设置 |
| `path_provider` | 本地文件路径 |
| `permission_handler` | 权限管理 |
| `package_info_plus` | 版本信息 |
| `device_info_plus` | 设备信息 |
| `url_launcher` | 打开协议、帮助和商店页面 |

### 调试和质量

| 包名 | 用途 |
| --- | --- |
| `logger` | 分级日志 |
| `talker_flutter` | 应用内日志和网络调试 |
| `flutter_lints` | 静态规则 |
| `mocktail` | 单元测试 mock |
| `golden_toolkit` | UI 快照测试 |
| `patrol` | 端到端测试 |

UI 风格建议使用 Material 3，并封装自己的 `AppTheme`、`AppColor`、`AppSpacing` 和 `AppTypography`。不要每个页面直接写颜色和尺寸。

## 5. 功能模块

| 模块 | 功能 | 优先级 |
| --- | --- | --- |
| 认证 | 注册、登录、验证码、第三方登录、退出、令牌刷新 | P0 |
| 家庭 | 创建家庭、邀请成员、角色权限、成员管理 | P0 |
| 儿童档案 | 姓名、年龄、头像、兴趣、内容等级、使用限制 | P0 |
| 设备绑定 | 扫码、蓝牙发现、配网、绑定、解绑、重命名、在线状态 | P0 |
| 家长策略 | 使用时长、禁用时段、内容权限、音量上限、单次时长 | P0 |
| 使用报告 | 对话次数、内容偏好、学习时长、使用趋势、异常提醒 | P1 |
| 远程留言 | 发送语音或文字留言、定时播放、已播放状态 | P1 |
| 内容管理 | 内容包、主题包、收藏、下载、删除、更新 | P1 |
| OTA | 版本查看、升级确认、进度、失败重试、版本回滚提示 | P1 |
| 安全 | 家长验证、敏感操作确认、隐私设置、数据删除 | P0 |
| 通知 | 设备离线、升级完成、儿童请求、内容更新 | P1 |

## 6. 与后台端的接口边界

家长 App 只访问后台服务：

- 使用 HTTPS REST API 处理账号、儿童、设备、策略、报告、内容和 OTA。
- 使用 WebSocket 或推送接收设备状态和异步任务结果。
- 不直接连接 ESP32。
- 不直接调用 AI 模型供应商。
- 不在客户端保存设备长期密钥。

建议的 API 前缀：

```text
/api/parent/v1/auth
/api/parent/v1/families
/api/parent/v1/children
/api/parent/v1/devices
/api/parent/v1/policies
/api/parent/v1/reports
/api/parent/v1/content
/api/parent/v1/ota
```

## 7. 可删除模块

以下功能应按 feature 独立：

```text
features/camera_settings/
features/display_settings/
features/cellular_settings/
features/motion_settings/
features/battery_settings/
```

当设备不支持对应能力时：

- 不显示入口。
- 不注册路由。
- 不请求相关接口。
- capability 变化时自动刷新菜单。

## 8. 首发范围

第一版只做：

1. 登录和家庭账号。
2. 扫码绑定设备。
3. 儿童档案。
4. 内容和时长策略。
5. 设备状态。
6. OTA 升级。
7. 基础使用报告。

摄像头、屏幕、4G、电池和动作设置只预留 capability 和路由接口。

## 9. 开发规范

- 页面使用 `feature` 命名，例如 `features/device_binding`。
- 每个 feature 包含 `presentation`、`application`、`domain`、`data`。
- 公共函数使用 `lower_snake_case`，类型使用 `PascalCase`。
- 所有 API 响应必须映射为领域模型，不在 UI 直接读取 JSON。
- 所有错误必须转成用户可理解的状态，不直接展示后端堆栈。
- 儿童隐私相关字段默认最小化采集。
- 删除某个 feature 时，只能影响其路由和依赖注册。
