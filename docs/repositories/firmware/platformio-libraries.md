# PlatformIO 库选型

本页是固件模块的依赖选型清单。所有依赖都必须锁定主版本，并只在对应模块启用时
参与构建。完整通用规则见
[第三方库选型](../technology-selection.md#4-platformio-与-esp32-s3)。

## 1. 框架和基础设施

| 场景 | 组件 | `lib_deps` 或 ESP-IDF 组件 | 适用模块 |
| --- | --- | --- | --- |
| ESP-IDF | `esp_wifi`、`esp_netif` | ESP-IDF 组件 | `network_manager` |
| MQTT/TLS | `mqtt` | ESP-IDF 组件 | `parent_link`、`ota_manager` |
| WebSocket | `esp_websocket_client` | ESP-IDF 组件 | `asr_client`、`llm_client`、`tts_client` |
| HTTP | `esp_http_client` | ESP-IDF 组件 | `cloud_auth`、`ota_download` |
| OTA | `esp_ota_ops` | ESP-IDF 组件 | `ota_manager` |
| 配置 | `nvs_flash` | ESP-IDF 组件 | `config_store` |
| 事件 | `esp_event` | ESP-IDF 组件 | `system_core` |
| 定时 | `esp_timer` | ESP-IDF 组件 | `alarm_reminder`、`usage_report` |
| 日志 | `esp_log` | ESP-IDF 组件 | `log_service` |
| JSON | ArduinoJson | `bblanchon/ArduinoJson @ ^7.4.2` | MQTT、HTTP、事件 |

## 2. 音频和语音

| 场景 | 组件或库 | 配置 | 适用模块 |
| --- | --- | --- | --- |
| 音频管线 | ESP-ADF | ESP-IDF 组件 | `audio_pipeline` |
| 音频设备 | `esp_codec_dev` | ESP-IDF 组件 | `audio_input`、`audio_output` |
| 唤醒 | ESP-SR | ESP-IDF 组件 | `voice_wake` |
| 回声消除 | SpeexDSP / ESP-SR | ESP-IDF 组件 | `far_field_audio` |
| 语音编解码 | Opus | ESP-IDF 组件 | `voice_session`、`video_session` |

## 3. 显示和触摸

| 场景 | 库 | `lib_deps` 或组件 | 适用模块 |
| --- | --- | --- | --- |
| TFT | LovyanGFX | `lovyan03/LovyanGFX @ ^1.2.7` | `display_ui` |
| TFT 备选 | TFT_eSPI | `bodmer/TFT_eSPI @ ^2.5.43` | `display_ui` |
| OLED | Adafruit SSD1306 | `adafruit/Adafruit SSD1306 @ ^2.5.15` | `display_ui` |
| 图形 | Adafruit GFX | `adafruit/Adafruit GFX Library @ ^1.11.11` | OLED 和基础图形 |
| GUI | LVGL | `lvgl/lvgl @ ^9.3.0` | `display_ui`、`expression_animation`、`point_read` |
| LVGL 接入 | `esp_lvgl_port` | ESP-IDF 组件 | `display_ui` |
| 触摸 | `esp_lcd_touch` | ESP-IDF 组件 | `touch_input`、`gesture_input` |

## 4. 输入、传感器和执行器

| 场景 | 库 | `lib_deps` | 适用模块 |
| --- | --- | --- | --- |
| 按键 | EasyButton | `easybtn/EasyButton @ ^2.0.4` | `button_input`、`factory_reset` |
| 舵机 | ESP32Servo | `madhephaestus/ESP32Servo @ ^3.0.6` | `servo_drive`、`motion_control` |
| 步进电机 | AccelStepper | `waspinator/AccelStepper @ ^1.64` | `motor_drive`、`walk_action` |
| 温度 | 内部传感器或 Adafruit BME280 | `adafruit/Adafruit BME280 Library @ ^2.3.0` | `temperature_monitor` |
| 湿度 | Adafruit DHT | `adafruit/DHT sensor library @ ^1.4.6` | 环境传感器模块 |
| 传感器基础 | Adafruit Unified Sensor | `adafruit/Adafruit Unified Sensor @ ^1.1.15` | 所有 Adafruit 传感器 |
| 摄像头 | `esp_camera` | ESP-IDF 组件 | `camera_capture`、`camera_vision` |
| 4G | TinyGSM | `vshymanskyy/TinyGSM @ ^0.12.0` | `cellular_4g` |
| 定位 | TinyGPSPlus | `mikalhart/TinyGPSPlus @ ^1.1.0` | `wearable_location` |

## 5. 依赖与模块绑定

示例：

```ini
[env:esp32-s3-n16r8]
platform = espressif32
board = esp32-s3-n16r8
framework = espidf
lib_deps =
  bblanchon/ArduinoJson @ ^7.4.2
  lovyan03/LovyanGFX @ ^1.2.7
  lvgl/lvgl @ ^9.3.0
  easybtn/EasyButton @ ^2.0.4
```

当模块关闭时，对应的 `lib_deps`、CMake 依赖和源码都必须从构建中移除。不能在
基础 Profile 中因为依赖关系隐式链接摄像头、屏幕、4G、电池或动作库。

## 6. 选型建议

- ESP-IDF 官方组件优先，尤其是网络、TLS、MQTT、OTA 和日志。
- JSON 统一使用 ArduinoJson，不手写解析器。
- GUI 统一使用 LVGL，显示驱动优先 LovyanGFX。
- 按键统一使用 EasyButton，传感器优先 Adafruit 系列。
- 4G 和定位使用 TinyGSM、TinyGPSPlus，不自行实现 AT 命令和 NMEA 解析。
- 任何新依赖必须先验证 ESP32-S3、ESP-IDF 版本、PSRAM 和许可证。
