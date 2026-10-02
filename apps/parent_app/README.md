<div align="center">

<img src="../../assets/brand/sprout/brand_banner.png" alt="如此萌屋" width="560" />

</div>

# 如此萌屋家长端

面向家长和监护人的 Flutter 应用，是芽系列·初芽平台的主要家庭入口。

## 职责

- 登录、家庭创建和监护人身份验证。
- 儿童档案、年龄分级和家长策略管理。
- 游戏机绑定、在线状态和基础设备管理。
- 远程留言、内容管理和 OTA 更新入口。
- 使用报告、隐私授权和儿童数据删除入口。

应用只访问平台业务服务，不直接连接设备，也不保存模型供应商密钥。

## 技术栈

| 场景 | 选型 |
| --- | --- |
| 框架 | Flutter、Dart |
| 路由 | `go_router` |
| 状态管理 | `flutter_riverpod` |
| 网络 | `dio` |
| 安全存储 | `flutter_secure_storage` |
| 序列化 | `json_serializable`、`json_annotation` |
| 测试 | `flutter_test`、`flutter_lints` |

## 本地运行

```powershell
flutter pub get
flutter run -d android
```

构建可安装的 Android 包：

```powershell
flutter build apk --release
```

当前维护范围只有 Android 和 iOS。Linux、Windows、macOS 和 Web
不属于家长端交付目标，因此没有保留对应平台目录。

## 目录

```text
lib/app/         应用入口、路由和主题装配
lib/core/        配置、网络、存储、错误和共享 UI 文案
lib/features/    按业务能力拆分的功能模块
lib/shared/      跨功能复用的无业务组件
test/            单元测试和组件测试
assets/brand/    品牌展示资源；应用图标由 flutter_launcher_icons 生成
```

每个 feature 必须通过公开边界访问，删除无关功能时不得破坏其他 feature
的编译、测试和运行。

## 质量门禁

```powershell
flutter analyze
flutter test
```
