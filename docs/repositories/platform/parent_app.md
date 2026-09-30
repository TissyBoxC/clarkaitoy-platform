# 家长控制端代码文档

## 1. 定位

`apps/parent_app` 是面向家长和监护人的 Flutter 应用，负责账号、家庭、儿童档案、
设备绑定、家长策略、报告、留言、内容、OTA、隐私和能力设置。

该应用只访问 `device_platform`。它不直接连接 ESP32，不直接调用模型供应商，
不保存设备长期密钥。

## 2. 目标目录

```text
apps/parent_app/
  lib/
    app/
      app.dart
      router/
        app_router.dart
        feature_registry.dart
    core/
      config/
      error/
      network/
      storage/
      theme/
      observability/
    features/
      auth/
      family/
      child_profile/
      device_binding/
      parent_policy/
      usage_report/
      remote_message/
      content_management/
      ota/
      notifications/
      privacy/
      camera_settings/
      display_settings/
      touch_settings/
      cellular_settings/
      battery_settings/
      motion_settings/
      video_companion/
      wearable_settings/
      multi_device/
    shared/
      widgets/
      forms/
      models/
    l10n/
  test/
    unit/
    widget/
    integration/
```

每个 feature 统一使用以下结构：

```text
features/<feature_name>/
  presentation/       页面、组件、状态绑定
  application/        用例、控制器、状态
  domain/             领域模型、接口、错误
  data/               API、本地缓存、DTO 映射
  routes.dart         feature 路由声明
  providers.dart      feature 依赖注册
  README.md           职责、接口、删除说明
```

## 3. 技术选型

完整版本和理由见
[第三方库选型](../technology-selection.md#2-flutter)。

| 场景 | 默认选择 | 说明 |
| --- | --- | --- |
| 框架 | Flutter 3.x / Dart 3.x | 当前项目基线 |
| 状态管理 | `flutter_riverpod` | 默认选择，测试友好 |
| 路由 | `go_router` | 官方推荐声明式路由 |
| 网络 | `dio` | 拦截器、超时、取消和上传下载 |
| 模型 | `freezed` + `json_serializable` | 不可变模型和生成序列化 |
| 安全存储 | `flutter_secure_storage` | 令牌和设备敏感配置 |
| 设置存储 | `shared_preferences` | 非敏感用户设置 |
| 数据库 | `drift` | 离线报告、内容和历史 |
| 推送 | `firebase_messaging` | 设备状态和异步结果 |
| UI | GetWidget 或 shadcn_ui | 不手写重复基础组件 |
| 图表 | `fl_chart` | 使用报告和学习趋势 |

## 4. 功能模块

### 4.1 P0

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `auth` | 注册、登录、验证码、刷新令牌、会话恢复、退出 | `device_platform/auth` |
| `family` | 创建家庭、成员、邀请、角色、移除成员 | `device_platform/family` |
| `child_profile` | 姓名、年龄、头像、兴趣、内容等级、使用限制 | `device_platform/child` |
| `device_binding` | 扫码、蓝牙发现、配网、绑定、解绑、重命名、在线状态 | `device_platform/device` |
| `parent_policy` | 时长、禁用时段、内容权限、音量、单次时长 | `device_platform/parent_policy` |
| `usage_report` | 使用时长、对话次数、内容偏好、异常提醒 | `device_platform/telemetry` |
| `ota` | 版本、升级、进度、失败重试、回滚提示 | `device_platform/ota` |
| `notifications` | 设备离线、升级、儿童请求、内容更新 | `device_platform/notification` |
| `privacy` | 家长验证、数据导出、数据删除、敏感操作确认 | `device_platform/auth`、`audit` |

### 4.2 P1

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `remote_message` | 语音或文字留言、定时播放、播放状态 | `device_platform/notification` |
| `content_management` | 内容包、主题包、收藏、下载、删除、更新 | `device_platform/content` |
| `companionship` | 成长记录、习惯提醒、鼓励设置、长记忆授权 | `device_platform/companion` |
| `english_learning` | 单词、句型、口语陪练、发音反馈报告 | `device_platform/learning` |
| `camera_settings` | 拍照问答、识物、绘本识别、图片隐私设置 | `device_platform/vision` |
| `diagnostics` | 网络质量、温度、错误报告和售后入口 | `device_platform/telemetry` |

### 4.3 P2

| 模块 | 页面/能力 | 依赖 |
| --- | --- | --- |
| `display_settings` | 亮度、色温、护眼、动画和视频设置 | `device_platform/display` |
| `touch_settings` | 触摸、手势和反馈设置 | `device_platform/input` |
| `cellular_settings` | eSIM、流量、漫游和远程唤醒 | `device_platform/connectivity` |
| `battery_settings` | 电量、充电、低功耗和休眠策略 | `device_platform/power` |
| `motion_settings` | 动作、跳舞、行走和停止控制 | `device_platform/motion` |
| `video_companion` | 视频通话、远程陪伴和通话记录 | `voice_gateway/video` |
| `wearable_settings` | 定位、电话、电子围栏、SOS | `device_platform/wearable` |
| `multi_device` | 家庭联动、设备组网、内容同步 | `device_platform/device_mesh` |

## 5. Capability 驱动

家长端不得根据型号硬编码菜单。设备 capability 来自设备接口或共享契约：

```dart
final capabilities = ref.watch(deviceCapabilitiesProvider(deviceId));

if (capabilities.contains(DeviceCapability.camera)) {
  // 注册 camera_settings 路由和入口。
}
```

规则：

- capability 不存在时，不显示入口、不注册路由、不请求接口。
- capability 变化时刷新首页、设置页和快捷入口。
- 功能模块关闭时，只移除模块注册，不修改其他 feature。
- feature 必须通过 `FeatureRegistry` 注册，不在全局路由文件中散落判断。

## 6. 接口边界

| 通道 | 用途 |
| --- | --- |
| HTTPS REST | 账号、家庭、儿童、设备、策略、报告、内容、OTA |
| WebSocket | 设备状态和异步任务进度 |
| Push | 离线、升级、儿童请求和内容更新 |

接口前缀：

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

API 响应必须映射到领域模型。UI 不直接读取 JSON，不展示后端堆栈。

## 7. 离线与错误

- `dio` 拦截器负责认证、重试、超时和错误映射。
- 网络错误、权限错误和业务错误在 `core/error` 统一分类。
- 离线状态下允许读取缓存报告、内容列表和设备状态。
- 敏感请求失败不得自动重放，必须由用户确认。
- 令牌刷新失败时清空安全存储并回到登录。

## 8. 测试

| 层级 | 内容 |
| --- | --- |
| 单元测试 | 状态、用例、DTO 映射、错误分类 |
| Widget 测试 | 表单、空状态、错误状态和 capability 菜单 |
| 集成测试 | 登录、绑定、策略、OTA、留言和隐私流程 |
| Golden 测试 | 关键页面和移动端宽度 |

每个 feature 至少包含 application、data 和 presentation 层测试。删除 feature
后，运行 `flutter analyze`、`flutter test` 和受影响路由测试。

## 9. 删除方法

1. 删除 `features/<feature_name>/`。
2. 从 `feature_registry.dart` 移除注册。
3. 从 `pubspec.yaml` 移除 feature 专属依赖。
4. 删除对应路由、权限、通知和契约消费者。
5. 运行 Flutter 分析和测试。

删除只允许影响 feature 注册和依赖声明，不修改无关 feature。
