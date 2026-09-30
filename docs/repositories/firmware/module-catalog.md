# 固件模块目录

## 1. 使用规则

每个模块都必须有独立 Kconfig 开关、CMake 条件编译、公开头文件、实现、测试和
README。模块名使用 `lower_snake_case`，公共函数使用
`<module>_<action>_<object>()`。

开关命名统一为：

```text
CONFIG_FEATURE_<MODULE_NAME>
```

## 2. 系统与基础模块

| 模块 | 优先级 | 职责 | 依赖 |
| --- | --- | --- | --- |
| `system_core` | P0 | 启动、版本、事件循环、任务和错误恢复 | ESP-IDF 基础组件 |
| `module_registry` | P0 | 注册、初始化、启动、停止模块 | `system_core` |
| `version_info` | P0 | 固件、硬件、协议和内容版本 | `system_core` |
| `error_recovery` | P0 | 模块失败隔离、重启、降级和恢复 | `system_core` |
| `config_store` | P0 | NVS 配置、设备身份和能力集合 | `nvs_flash` |
| `transport_security` | P0 | TLS、证书、设备令牌和请求校验 | `esp_tls`、`mbedtls` |
| `privacy_guard` | P0 | 隐私开关、数据最小化、授权检查 | `config_store` |
| `content_filter` | P0 | 本地内容和 AI 输出过滤 | `content_library` |

## 3. 音频和语音模块

| 模块 | 优先级 | 职责 | 依赖 |
| --- | --- | --- | --- |
| `audio_input` | P0 | 麦克风采集、增益、降噪和帧输出 | `audio_pipeline` |
| `audio_output` | P0 | 扬声器播放、音量、混音和静音 | `audio_pipeline` |
| `audio_pipeline` | P0 | 音频管线、采样率、编码和缓冲 | ESP-ADF 或 codec adapter |
| `playback_queue` | P0 | 播放队列、优先级、打断和恢复 | `audio_output` |
| `voice_wake` | P0 | 唤醒词检测、误唤醒控制和唤醒事件 | ESP-SR |
| `wake_feedback` | P0 | 唤醒提示音、LED 或屏幕反馈 | `audio_output`、`led_indicator` |
| `voice_session` | P0 | 对话状态机、录音、上传、打断和超时 | `audio_input`、网络、ASR、LLM、TTS |
| `conversation_context` | P0 | 多轮上下文、会话 ID 和过期策略 | `voice_session` |
| `child_prompt_profile` | P0 | 儿童年龄、语气、主题和安全提示词 | `parent_policy` |
| `asr_client` | P0 | ASR 网络客户端和重连 | `network_manager`、WebSocket |
| `llm_client` | P0 | AI 网关调用和流式响应 | `network_manager`、WebSocket |
| `tts_client` | P0 | TTS 请求、音频流和播放控制 | `network_manager`、`audio_output` |
| `full_duplex_voice` | P1 | 全双工对话和打断 | `voice_session` |
| `far_field_audio` | P1 | 远场拾音、回声消除和噪声抑制 | `audio_input`、SpeexDSP |
| `continuous_conversation` | P1 | 唤醒后连续会话和自动结束 | `voice_session` |

## 4. 内容模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `content_library` | P0 | 本地内容索引、播放和缓存 |
| `content_story` | P0 | 故事内容组织和播放 |
| `content_nursery_rhyme` | P0 | 儿歌内容组织和播放 |
| `content_poetry` | P0 | 古诗内容组织和播放 |
| `content_english` | P0 | 英语内容组织和播放 |
| `content_encyclopedia` | P0 | 百科内容组织和播放 |
| `content_bedtime` | P0 | 睡前内容、定时和音量策略 |
| `content_age_tier` | P0 | 分龄内容筛选和推荐 |
| `content_package_manager` | P1 | 内容包安装、更新和卸载 |
| `theme_package` | P1 | 主题包、音色和界面资源 |
| `content_update` | P1 | 增量更新、校验和断点续传 |
| `favorites` | P1 | 收藏、取消收藏和快捷播放 |
| `playback_history` | P1 | 播放记录、续播和使用统计 |
| `learning_visualization` | P2 | 点读、学习卡片和可视化反馈 |

## 5. 家长、陪伴与学习模块

| 模块 | 优先级 | 职责 | 依赖 |
| --- | --- | --- | --- |
| `parent_link` | P0 | 家长策略、远程消息和状态同步 | MQTT、网络 |
| `parent_policy` | P0 | 时长、禁用时段、内容等级和音量 | `config_store` |
| `usage_report` | P0 | 使用时长、对话次数和内容偏好 | 本地存储、MQTT |
| `companionship` | P1 | 陪伴模式、情绪回应和长期记忆开关 | `voice_session` |
| `emotion_response` | P1 | 情绪识别结果和回应策略 | `companionship` |
| `encouragement` | P1 | 鼓励语、成就和正向反馈 | `companionship` |
| `growth_record` | P1 | 成长记录和本地事件 | `usage_report` |
| `habit_reminder` | P1 | 习惯提醒和时间计划 | `alarm_reminder` |
| `alarm_reminder` | P1 | 闹钟、定时提醒和日程 | `time_sync` |
| `long_term_memory` | P1 | 经授权的长期陪伴记忆 | `privacy_guard` |
| `english_learning` | P1 | 英语学习入口和课程状态 | `content_library` |
| `word_practice` | P1 | 单词练习和复习 | `english_learning` |
| `sentence_practice` | P1 | 句型练习和对话 | `english_learning` |
| `speaking_practice` | P1 | 口语练习和录音 | `voice_session` |
| `pronunciation_feedback` | P1 | 发音反馈和评分 | `speaking_practice` |

## 6. 远程、兜底与诊断模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `remote_message` | P1 | 家长文字或语音留言 |
| `remote_playback` | P1 | 远程点播和播放控制 |
| `device_status` | P1 | 在线、电量、网络和会话状态 |
| `content_recommendation` | P1 | 家长推荐内容接收 |
| `offline_fallback` | P1 | 断网检测和降级 |
| `offline_story` | P1 | 离线故事播放 |
| `local_commands` | P1 | 本地唤醒、播放、停止和音量命令 |
| `cache_playback` | P1 | 缓存内容播放和续播 |
| `system_diagnostics` | P1 | 日志、崩溃、网络和温度诊断 |
| `log_service` | P1 | 分级日志、滚动和脱敏 |
| `crash_report` | P1 | 崩溃记录和上报 |
| `network_quality` | P1 | 延迟、丢包和信号质量 |
| `temperature_monitor` | P1 | 芯片和硬件温度保护 |

## 7. 摄像头与视觉模块

| 模块 | 优先级 | 职责 | 依赖 |
| --- | --- | --- | --- |
| `camera_vision` | P1 | 视觉能力总入口和任务调度 | `camera_capture` |
| `camera_capture` | P1 | 摄像头采集、曝光和图片格式 | `esp_camera` |
| `object_recognition` | P1 | 识物任务和结果 | `camera_vision` |
| `picture_book_recognition` | P1 | 绘本识别和页码 | `camera_vision` |
| `photo_qa` | P1 | 拍照问答和 AI 上下文 | `camera_vision`、`voice_session` |
| `image_privacy` | P1 | 图片本地处理、授权、脱敏和删除 | `privacy_guard` |

## 8. 显示与交互模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `display_ui` | P2 | 屏幕、页面、状态和布局 |
| `expression_animation` | P2 | 表情和动画 |
| `point_read` | P2 | 点读和内容定位 |
| `video_playback` | P2 | 视频播放和资源管理 |
| `eye_care_display` | P2 | 亮度、色温、距离和护眼提醒 |
| `touch_input` | P2 | 触摸坐标和点击 |
| `gesture_input` | P2 | 滑动、长按和组合手势 |
| `haptic_feedback` | P2 | 振动和触觉反馈 |
| `led_indicator` | P0/P2 | 状态灯和灯光反馈 |
| `button_input` | P0 | 按键、组合键和长按 |
| `device_interaction` | P0 | 交互事件编排和状态反馈 |
| `prompt_tone` | P0 | 提示音和系统反馈 |
| `volume_control` | P0 | 音量、静音和音量限制 |
| `factory_reset` | P0 | 复位、清除配置和恢复出厂 |
| `bluetooth_audio` | P1 | 蓝牙音箱模式和音频输入 |

## 9. 网络、电源与差异化模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `network_manager` | P0 | WiFi、MQTT、WebSocket、重连和网络切换 |
| `wifi_provisioning` | P0 | 配网、热点、二维码和凭证保存 |
| `time_sync` | P0 | NTP、时区和时间校准 |
| `cloud_auth` | P0 | 设备身份、令牌和设备鉴权 |
| `ota_manager` | P0 | OTA 调度和升级状态 |
| `ota_download` | P0 | 固件下载、断点和重试 |
| `ota_validate` | P0 | 签名、哈希和版本校验 |
| `ota_rollback` | P0 | 分区回滚和失败恢复 |
| `cellular_4g` | P2 | 4G 网络、APN 和切换 |
| `esim_management` | P2 | eSIM 配置和激活 |
| `data_usage_management` | P2 | 流量统计、限额和提醒 |
| `remote_wake` | P2 | 蜂窝或网络远程唤醒 |
| `battery_power` | P2 | 电池、充电和电源状态 |
| `charge_management` | P2 | 充电保护和充电策略 |
| `battery_estimation` | P2 | 电量和续航估算 |
| `low_power` | P2 | 低功耗模式和任务降频 |
| `deep_sleep` | P2 | 深度休眠和唤醒源 |

## 10. 动作、拟人、视频和穿戴模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `motion_control` | P2 | 动作总入口、安全限制和停止 |
| `motor_drive` | P2 | 直流电机或轮组驱动 |
| `servo_drive` | P2 | 舵机控制 |
| `walk_action` | P2 | 行走和转向动作 |
| `dance_action` | P2 | 跳舞和节奏动作 |
| `tail_action` | P2 | 尾巴、耳朵等形态动作 |
| `form_factor_profile` | P2 | 机器狗、机器猫、毛绒等形态能力 |
| `video_call` | P2 | 视频通话和状态 |
| `remote_companionship` | P2 | 远程陪伴和共享会话 |
| `video_session` | P2 | 视频媒体会话、编解码和同步 |
| `wearable_location` | P2 | 定位、轨迹和位置上报 |
| `phone_call` | P2 | 语音电话和联系人 |
| `geofence` | P2 | 电子围栏和越界提醒 |
| `sos` | P2 | SOS 按钮、定位和通知 |

## 11. 多设备模块

| 模块 | 优先级 | 职责 |
| --- | --- | --- |
| `family_linkage` | P2 | 家庭设备联动和场景 |
| `device_mesh` | P2 | 设备发现、组网和消息路由 |
| `content_sync` | P2 | 多设备内容同步和版本一致 |

## 12. 模块验收

每个模块必须提供：

- 公开头文件和使用示例。
- 独立 Kconfig 开关。
- 独立 CMake 条件编译。
- 正常、失败、超时、重连和断电恢复测试。
- 资源占用和任务优先级说明。
- 模块删除步骤。
- 与其他模块的公开接口和依赖说明。

## 13. 发行配置

建议至少提供以下构建配置：

| Profile | 包含能力 |
| --- | --- |
| `audio_lite` | 音频、WiFi、内容、家长控制、OTA、诊断 |
| `vision_standard` | `audio_lite` + 摄像头和图片隐私 |
| `screen_pro` | `vision_standard` + 屏幕、触摸、护眼和可视化 |
| `motion_plus` | `screen_pro` + 动作、拟人和多设备 |
| `connected_plus` | `motion_plus` + 4G、电池、视频和穿戴 |

所有 Profile 共享同一套模块源码和测试，只通过 Kconfig 和 CMake 选择模块。
