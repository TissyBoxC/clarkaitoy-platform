# 固件构建与发布

## 1. 构建目标

```text
芯片：ESP32-S3 N16R8
Flash：16 MB
PSRAM：8 MB OPI PSRAM
框架：ESP-IDF
构建系统：PlatformIO
构建目录：bgen/
```

所有构建输出、中间文件、镜像、测试报告和烧录产物都放入 `bgen/`，不得提交到
Git。

## 2. 基础命令

```powershell
Set-Location firmware
platformio run -e esp32-s3-n16r8
platformio run -e esp32-s3-n16r8 -t upload
platformio device monitor -b 115200
```

## 3. 配置开关

每个模块都有独立开关：

```text
CONFIG_FEATURE_SYSTEM_CORE
CONFIG_FEATURE_MODULE_REGISTRY
CONFIG_FEATURE_AUDIO_INPUT
CONFIG_FEATURE_AUDIO_OUTPUT
CONFIG_FEATURE_VOICE_WAKE
CONFIG_FEATURE_VOICE_SESSION
CONFIG_FEATURE_ASR_CLIENT
CONFIG_FEATURE_LLM_CLIENT
CONFIG_FEATURE_TTS_CLIENT
CONFIG_FEATURE_CONTENT_LIBRARY
CONFIG_FEATURE_PARENT_LINK
CONFIG_FEATURE_NETWORK_MANAGER
CONFIG_FEATURE_OTA_MANAGER
CONFIG_FEATURE_CAMERA_VISION
CONFIG_FEATURE_DISPLAY_UI
CONFIG_FEATURE_TOUCH_INPUT
CONFIG_FEATURE_LED_INDICATOR
CONFIG_FEATURE_BATTERY_POWER
CONFIG_FEATURE_CELLULAR_4G
CONFIG_FEATURE_MOTION_CONTROL
```

完整模块开关见 [固件模块目录](module-catalog.md)。

## 4. 条件编译规则

每个组件 `CMakeLists.txt` 应遵循：

```cmake
idf_component_register(
    SRCS "src/<module_name>.c"
    INCLUDE_DIRS "include"
    REQUIRES <public_dependencies>
)
```

组合根只在开关打开时调用模块注册：

```c
void app_register_modules(void)
{
#if CONFIG_FEATURE_AUDIO_INPUT
    module_registry_add(audio_input_module_get());
#endif
#if CONFIG_FEATURE_CAMERA_VISION
    module_registry_add(camera_vision_module_get());
#endif
#if CONFIG_FEATURE_DISPLAY_UI
    module_registry_add(display_ui_module_get());
#endif
}
```

禁止用运行时 `if` 代替 `#if` 处理可选硬件模块，否则关闭模块后仍会链接和占用
资源。

## 5. 发行 Profile

| Profile | 开关策略 | 目标 |
| --- | --- | --- |
| `audio_lite` | 基础音频、WiFi、内容、家长控制、OTA | 首版验证 |
| `vision_standard` | `audio_lite` + 摄像头和图片隐私 | 视觉差异化 |
| `screen_pro` | `vision_standard` + 屏幕、触摸、护眼 | 可视化学习 |
| `motion_plus` | `screen_pro` + 动作、拟人和多设备 | 玩具和陪伴 |
| `connected_plus` | `motion_plus` + 4G、电池、视频、穿戴 | 户外和移动场景 |

每个 Profile 都必须通过：

```text
平台无关逻辑测试
目标板编译
镜像大小检查
启动和模块初始化检查
核心链路回归
```

## 6. 依赖管理

- ESP-IDF 原生能力使用官方组件。
- 第三方库使用 [PlatformIO 库选型](platformio-libraries.md) 中经过验证的版本。
- 每个组件的依赖写在 `CMakeLists.txt` 或 `idf_component.yml`。
- 模块关闭时，不得下载或链接其专属库。
- 依赖必须锁定版本，禁止浮动 latest。

## 7. 分区和资源

- 分区表支持 OTA A/B 分区、NVS、内容缓存和诊断数据。
- 内容缓存与固件分区隔离，删除固件模块不得误删用户内容。
- 图片、音频和模型资源必须有大小预算。
- OTA 镜像、回滚镜像和内容包大小必须在 CI 中检查。

## 8. 发布流程

1. 运行全部模块单元测试。
2. 编译所有发行 Profile。
3. 检查固件、资源和分区大小。
4. 运行核心闭环和可选模块冒烟测试。
5. 生成版本清单：固件版本、模块开关、内容版本和兼容协议版本。
6. 上传固件包和签名。
7. 在平台仓库更新 `workspace.lock.yaml` 的固件 revision。

## 9. 提交前检查

```powershell
git status --short --branch
git diff
git diff --cached
platformio run -e esp32-s3-n16r8
```

确认不提交 `bgen/`、`AGENTS.md`、本地证书、设备配置、日志和生成镜像。
