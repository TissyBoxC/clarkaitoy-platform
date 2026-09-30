# 游戏机本体开发文档

## 1. 项目定位

游戏机本体运行在 ESP32-S3 N16R8 上，是面向幼儿的 AI 早教陪伴设备。

它负责：

- 唤醒词检测
- 麦克风采集和扬声器播放
- AI 语音对话
- 故事、儿歌、英语和本地内容播放
- 按键、LED、触摸、屏幕等交互
- 摄像头等可选视觉能力
- 家长策略执行
- WiFi、4G 等网络连接
- OTA 升级
- 设备状态和错误诊断

它不负责：

- 长期保存儿童档案
- 内容运营
- 家长账号管理
- AI 供应商账号池
- 复杂计费和配额

## 2. 在项目中的组成

```text
firmware/
  esp32-s3-n16r8/
    main/
    components/
      system_core/
      audio_io/
      voice_wake/
      voice_session/
      asr_client/
      llm_client/
      tts_client/
      content_library/
      parent_link/
      network_manager/
      ota_manager/
      privacy_guard/
      system_diagnostics/
      camera_vision/
      display_ui/
      touch_input/
      led_indicator/
      battery_power/
      cellular_4g/
      motion_control/
    boards/
    partitions.csv
    platformio.ini
    sdkconfig.defaults
```

`main` 只做组合和启动，不写业务逻辑。

## 3. 技术栈

| 层级 | 推荐技术 |
| --- | --- |
| 芯片 | ESP32-S3 N16R8 |
| Flash | 16 MB |
| PSRAM | 8 MB OPI PSRAM |
| 开发框架 | ESP-IDF |
| 构建工具 | PlatformIO |
| 语言 | C |
| 构建目录 | `bgen/` |
| 任务模型 | FreeRTOS |
| 网络 | WiFi、MQTT/TLS、WebSocket |
| 音频 | ESP-ADF 或独立 audio pipeline |
| 唤醒 | ESP-SR / WakeNet |
| 语音识别 | 后台 ASR 或本地轻量识别 |
| 大模型 | 通过后台 AI 网关调用 `sub2api` |
| 语音合成 | 后台 TTS 或设备端轻量 TTS |
| OTA | ESP-IDF OTA |
| 配置 | Kconfig + `sdkconfig.defaults` |

## 4. 推荐实用工具包

### ESP-IDF 官方组件

| 组件 | 用途 |
| --- | --- |
| `esp_wifi` | WiFi |
| `esp_netif` | 网络接口 |
| `mqtt` | MQTT/TLS |
| `esp_websocket_client` | WebSocket |
| `esp_http_client` | HTTPS |
| `esp_ota_ops` | OTA |
| `nvs_flash` | 设备配置和密钥 |
| `esp_timer` | 定时和调度 |
| `esp_event` | 模块事件 |
| `esp_console` | 调试命令行 |
| `esp_system` | 重启、崩溃和诊断 |
| `esp_camera` | 摄像头 |
| `esp_lcd` | 屏幕 |
| `esp_lvgl_port` | LVGL 接入 |

### 音频和语音

| 组件 | 用途 |
| --- | --- |
| ESP-ADF | 音频管线、编解码、播放和录音 |
| ESP-SR | 唤醒词和语音前端 |
| `esp_codec_dev` | 音频编解码器抽象 |
| Opus | 低带宽语音压缩 |
| SpeexDSP | 回声消除和降噪 |

### UI 和显示

| 组件 | 用途 |
| --- | --- |
| LVGL | 嵌入式 UI |
| `esp_lvgl_port` | LVGL 与 ESP-IDF 集成 |
| `lv_font_*` | 中文和儿童字体 |
| PNG/JPEG 解码组件 | 图片资源 |

### 开发辅助

| 工具 | 用途 |
| --- | --- |
| PlatformIO CLI | 构建和上传 |
| `idf.py monitor` | 串口日志 |
| `esptool.py` | 烧录和芯片信息 |
| `clang-format` | C 代码格式化 |
| `cppcheck` | 静态分析 |
| Wireshark | 网络调试 |
| MQTTX | MQTT 调试 |

## 5. 功能模块

| 模块 | 负责功能 | 必选 |
| --- | --- | --- |
| system_core | 启动、任务、事件、配置、错误恢复 | 是 |
| audio_io | 麦克风、扬声器、音频管线 | 是 |
| voice_wake | 唤醒词和语音前端 | 是 |
| voice_session | 对话状态机、语音打断、超时 | 是 |
| asr_client | 语音识别服务连接 | 是 |
| llm_client | AI 网关调用 | 是 |
| tts_client | 语音合成和播放 | 是 |
| content_library | 本地内容、缓存和播放列表 | 是 |
| parent_link | 家长策略和远程消息 | 是 |
| network_manager | WiFi、MQTT、WebSocket、重连 | 是 |
| ota_manager | 版本、下载、校验、升级、回滚 | 是 |
| privacy_guard | 隐私开关、数据最小化、安全启动 | 是 |
| system_diagnostics | 日志、崩溃、网络和温度状态 | 是 |
| camera_vision | 拍照、识物、绘本识别 | 否 |
| display_ui | 表情、动画、点读和屏幕 UI | 否 |
| touch_input | 触摸交互 | 否 |
| led_indicator | 状态灯和灯光反馈 | 否 |
| battery_power | 电池、充电和低功耗 | 否 |
| cellular_4g | 4G 网络 | 否 |
| motion_control | 电机、舵机和动作 | 否 |

## 6. 语音对话链路

推荐链路：

```text
唤醒词
  -> 录音
  -> 音频前处理
  -> WebSocket 上传
  -> 后台 ASR
  -> 后台内容安全
  -> 后台 AI 网关
  -> sub2api
  -> 后台 TTS
  -> 设备播放
```

设备不直接持有模型供应商密钥，只持有设备身份和设备令牌。

## 7. 构建开关

```text
CONFIG_FEATURE_CAMERA
CONFIG_FEATURE_DISPLAY
CONFIG_FEATURE_TOUCH
CONFIG_FEATURE_LED
CONFIG_FEATURE_BATTERY
CONFIG_FEATURE_CELLULAR_4G
CONFIG_FEATURE_MOTION
```

删除模块时的规则：

1. 关闭 Kconfig 开关。
2. 从 `main` 的组合根移除初始化。
3. 从 `CMakeLists.txt` 移除依赖。
4. 不修改其他业务模块。
5. 重新构建 `bgen` 并验证固件大小。

## 8. 与后台服务的接口

| 通道 | 用途 |
| --- | --- |
| MQTT/TLS | 心跳、状态、策略、命令、OTA |
| WebSocket | 流式语音、实时对话 |
| HTTPS | 鉴权、内容下载、固件下载 |

设备启动后：

1. 加载 NVS 配置。
2. 建立 WiFi 或 4G 网络。
3. 连接后台。
4. 上报设备、固件和能力。
5. 拉取家长策略和内容包。
6. 进入待唤醒状态。

## 9. 首发范围

第一版只启用：

```text
audio_io
voice_wake
voice_session
asr_client
llm_client
tts_client
content_library
parent_link
network_manager
ota_manager
privacy_guard
system_diagnostics
led_indicator
```

摄像头、屏幕、触摸、4G、电池和动作只预留接口。

## 10. 开发规范

- 模块名、文件名和方法名使用 `lower_snake_case`。
- 公共函数使用 `<module>_<action>_<object>()`。
- 类型使用 `PascalCase`，宏使用 `UPPER_SNAKE_CASE`。
- 公共 API 必须写参数、返回值、失败行为和线程约束。
- 业务模块不能直接操作 GPIO、I2C、SPI 或摄像头寄存器。
- 硬件访问必须封装在 HAL 或适配模块中。
- 跨模块依赖只使用公开头文件。
- 所有可选模块必须在不修改其他业务模块的前提下删除。
- 每次修改后运行 `platformio run -e esp32-s3-n16r8`。
