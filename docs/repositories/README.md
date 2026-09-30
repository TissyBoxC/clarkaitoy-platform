# 仓库代码文档索引

本目录按仓库和子项目拆分，描述各仓库负责的功能、目录结构、技术选型、模块边界、
接口协议、测试和发布方式。

## 文档结构

```text
docs/repositories/
  technology-selection.md
  platform/
    README.md
    parent_app.md
    admin_web.md
    device_platform.md
    voice_gateway.md
    contracts.md
  firmware/
    README.md
    module-catalog.md
    build-and-release.md
    platformio-libraries.md
  sub2api/
    README.md
    backend.md
    admin-frontend.md
```

## 仓库边界

| 仓库 | 文档入口 | 职责 |
| --- | --- | --- |
| `clarkaitoy-platform` | [平台仓库](platform/README.md) | 家长端、管理端、Go 服务和共享契约 |
| `clarkaitoy-firmware` | [固件仓库](firmware/README.md) | ESP32-S3 N16R8 设备端和全部可裁剪模块 |
| `clarkaitoy-sub2api-fork` | [sub2api fork](sub2api/README.md) | 多模型网关、儿童策略和用量审计 |

## 跨仓库原则

- 功能范围以 [功能覆盖矩阵](../features/feature-coverage.md) 为准。
- 库选型以 [第三方库选型](technology-selection.md) 为准。
- 跨端协议以 `packages/contracts` 为准，仓库内部协议保留在各自仓库。
- 每个功能模块都要有明确的拥有者、公开接口、测试和删除方式。
- 平台仓库文档可以描述外部仓库，但外部仓库的源码必须在自己的仓库中提交。
