# 功能覆盖矩阵

## 1. 使用说明

本矩阵把市场调研中的基础层、主流层、差异化层以及 P0、P1、P2 功能映射到仓库、
服务模块、固件模块和构建开关。

- 所有条目都是实现范围，不能因优先级较低而删除。
- 固件开关用于选择发行配置，不代表删除功能源码。
- 平台服务可以按模块组合部署，但模块代码和测试必须保留。
- 每个固件模块都必须有独立 Kconfig 开关、CMake 条件编译和模块测试。

## 2. P0 功能覆盖

| ID | 功能 | 固件模块 | 平台模块 | 验收标准 |
| --- | --- | --- | --- | --- |
| P0-01 | 系统启动、版本和错误恢复 | `system_core`、`module_registry`、`version_info`、`error_recovery` | 无 | 设备启动、读取版本、单项模块失败可诊断且不影响无关模块 |
| P0-02 | 音频输入 | `audio_input`、`audio_pipeline` | `voice_gateway` | 可采集、增益控制、降噪和回声消除，音频帧格式可验证 |
| P0-03 | 音频输出 | `audio_output`、`playback_queue` | `voice_gateway` | 可播放 TTS 和本地内容，支持音量、队列、打断和恢复 |
| P0-04 | 唤醒 | `voice_wake`、`wake_feedback` | `voice_gateway` | 可配置唤醒词、误唤醒控制和唤醒反馈 |
| P0-05 | 语音会话 | `voice_session`、`asr_client`、`llm_client`、`tts_client` | `voice_gateway`、`sub2api_fork` | 录音、ASR、LLM、TTS、打断和超时闭环可运行 |
| P0-06 | AI 对话 | `conversation_context`、`child_prompt_profile` | `voice_gateway`、`sub2api_fork` | 多轮上下文、儿童提示词、超时和降级可验证 |
| P0-07 | 内容 | `content_library`、`content_story`、`content_nursery_rhyme`、`content_poetry`、`content_english`、`content_encyclopedia`、`content_bedtime`、`content_age_tier` | `device_platform`、`admin_web`、`parent_app` | 故事、儿歌、古诗、英语、百科、睡前和分龄内容可下发与播放 |
| P0-08 | 联网 | `network_manager`、`wifi_provisioning`、`time_sync`、`cloud_auth` | `device_platform` | 配网、重连、时间同步和云端鉴权可恢复 |
| P0-09 | 家长控制 | `parent_link`、`parent_policy`、`usage_report` | `device_platform`、`parent_app`、`admin_web` | 绑定、内容等级、时长、禁用时段和报告可闭环 |
| P0-10 | 安全与隐私 | `privacy_guard`、`content_filter`、`transport_security` | `device_platform`、`voice_gateway`、`sub2api_fork` | TLS、数据最小化、内容过滤、日志脱敏和敏感数据删除可验证 |
| P0-11 | OTA | `ota_manager`、`ota_download`、`ota_validate`、`ota_rollback` | `device_platform`、`admin_web`、`parent_app` | 下载、校验、安装、回滚和版本统计可验证 |
| P0-12 | 设备交互 | `device_interaction`、`button_input`、`led_indicator`、`prompt_tone`、`volume_control`、`factory_reset` | `parent_app`、`admin_web` | 按键、LED、提示音、音量、复位和状态反馈可用 |

## 3. P1 功能覆盖

| ID | 功能 | 固件模块 | 平台模块 | 验收标准 |
| --- | --- | --- | --- | --- |
| P1-01 | 语音增强 | `full_duplex_voice`、`far_field_audio`、`continuous_conversation` | `voice_gateway` | 全双工、远场拾音和唤醒后连续会话可用 |
| P1-02 | 内容运营 | `content_package_manager`、`theme_package`、`content_update`、`favorites`、`playback_history` | `device_platform`、`admin_web` | 内容包、主题包、更新、收藏和播放历史可闭环 |
| P1-03 | 陪伴 | `companionship`、`emotion_response`、`encouragement`、`growth_record`、`habit_reminder` | `device_platform`、`voice_gateway`、`parent_app` | 情绪回应、鼓励、成长记录和习惯提醒可验证 |
| P1-04 | 英语 | `english_learning`、`word_practice`、`sentence_practice`、`speaking_practice`、`pronunciation_feedback` | `device_platform`、`voice_gateway`、`parent_app` | 单词、句型、口语练习和发音反馈可运行 |
| P1-05 | 家长端 | `remote_message`、`remote_playback`、`device_status`、`content_recommendation` | `parent_app`、`device_platform` | 留言、点播、状态和推荐可下发到设备 |
| P1-06 | 本地兜底 | `offline_fallback`、`offline_story`、`local_commands`、`cache_playback` | `device_platform` | 断网故事、本地指令和缓存播放可用 |
| P1-07 | 诊断 | `system_diagnostics`、`log_service`、`crash_report`、`network_quality`、`temperature_monitor` | `device_platform`、`admin_web` | 日志、崩溃、网络质量、温度和服务告警可查询 |
| P1-08 | 摄像头 | `camera_vision`、`camera_capture`、`object_recognition`、`picture_book_recognition`、`photo_qa` | `voice_gateway`、`device_platform` | 拍照、识物、绘本识别和拍照问答可闭环 |

## 4. P2 功能覆盖

| ID | 功能 | 固件模块 | 平台模块 | 验收标准 |
| --- | --- | --- | --- | --- |
| P2-01 | 屏幕 | `display_ui`、`expression_animation`、`point_read`、`video_playback` | `device_platform`、`admin_web` | 表情、动画、点读和视频播放可用 |
| P2-02 | 护眼屏幕 | `eye_care_display` | `parent_app`、`device_platform` | 亮度、色温、时长提醒和护眼策略可配置 |
| P2-03 | 触摸 | `touch_input`、`gesture_input`、`haptic_feedback` | `parent_app`、`device_platform` | 点击、滑动、手势和触觉反馈可验证 |
| P2-04 | 4G | `cellular_4g`、`esim_management`、`data_usage_management`、`remote_wake` | `device_platform`、`admin_web` | eSIM、流量、远程唤醒和联网切换可闭环 |
| P2-05 | 电池 | `battery_power`、`charge_management`、`battery_estimation`、`low_power`、`deep_sleep` | `device_platform`、`parent_app` | 充电、电量估算、低功耗和休眠可验证 |
| P2-06 | 动作 | `motion_control`、`motor_drive`、`servo_drive`、`walk_action`、`dance_action`、`tail_action` | `device_platform`、`admin_web` | 电机、舵机和动作编排可配置、可停止 |
| P2-07 | 拟人形态 | `form_factor_profile` | `device_platform`、`admin_web` | 机器狗、机器猫和毛绒形态可切换能力集合 |
| P2-08 | 视频陪伴 | `video_call`、`remote_companionship` | `voice_gateway`、`device_platform`、`parent_app` | 视频会话、远程陪伴和通话状态可闭环 |
| P2-09 | 穿戴与定位 | `wearable_location`、`phone_call`、`geofence`、`sos` | `device_platform`、`parent_app`、`admin_web` | 定位、通话、电子围栏和 SOS 可闭环 |
| P2-10 | 多设备 | `family_linkage`、`device_mesh`、`content_sync` | `device_platform`、`parent_app` | 家庭联动、设备组网和内容同步可验证 |

## 5. 分层功能映射

### 5.1 基础层

| 调研功能 | 对应实现 |
| --- | --- |
| 开机和配网 | `system_core`、`wifi_provisioning`、`network_manager` |
| 语音唤醒 | `voice_wake`、`wake_feedback` |
| 录音和播放 | `audio_input`、`audio_output`、`audio_pipeline` |
| AI 对话 | `voice_session`、`asr_client`、`llm_client`、`tts_client` |
| 音量调节 | `volume_control` |
| 故事、儿歌、英语等基础内容 | `content_library` 及内容分类模块 |
| 断网提示和错误处理 | `network_manager`、`offline_fallback`、`error_recovery` |
| OTA 固件升级 | `ota_manager`、`ota_download`、`ota_validate`、`ota_rollback` |

### 5.2 主流层

| 调研功能 | 对应实现 |
| --- | --- |
| 多轮对话 | `conversation_context`、`voice_session` |
| 儿童化语音和语调 | `child_prompt_profile`、`tts_client` |
| 大模型切换或聚合 | `voice_gateway`、`sub2api_fork` |
| 英语口语陪练 | `english_learning` 系列模块 |
| 教材或分龄内容 | `content_age_tier`、`content_package_manager` |
| 定时提醒和闹钟 | `habit_reminder`、`alarm_reminder` |
| 家长控制 | `parent_link`、`parent_policy`、`usage_report` |
| 使用记录和使用时长 | `usage_report`、`device_platform` |
| 蓝牙音箱 | `bluetooth_audio` |

### 5.3 差异化层

| 调研功能 | 对应实现 |
| --- | --- |
| 情绪陪伴和长记忆 | `companionship`、`long_term_memory` |
| 摄像头识物或拍照问答 | `camera_vision` 系列模块 |
| 触屏和可视化学习 | `display_ui`、`touch_input`、`learning_visualization` |
| 护眼屏幕 | `eye_care_display` |
| 4G 独立联网 | `cellular_4g` 系列模块 |
| 机器狗、机器猫、毛绒等形态 | `form_factor_profile`、`motion_control` |
| 动作、触摸和灯光反馈 | `motion_control`、`touch_input`、`led_indicator` |
| 视频通话或远程陪伴 | `video_call`、`remote_companionship` |
| 定位和穿戴能力 | `wearable_location`、`phone_call`、`geofence`、`sos` |

## 6. 补充模块

以下模块没有单独出现在调研标题中，但为实现上述能力所必需，也必须模块化：

| 模块 | 用途 |
| --- | --- |
| `module_registry` | 统一注册、初始化和停止可选模块 |
| `config_store` | 保存设备配置和能力集合 |
| `playback_queue` | 管理音频播放优先级、打断和恢复 |
| `alarm_reminder` | 定时提醒、闹钟和日程 |
| `bluetooth_audio` | 蓝牙音箱模式 |
| `learning_visualization` | 触屏学习内容和可视化反馈 |
| `long_term_memory` | 经家长授权保存长期陪伴记忆 |
| `image_privacy` | 图片本地处理和上传授权 |
| `video_session` | 视频通话和远程陪伴的媒体会话 |

## 7. 实施约束

- 所有功能必须先在共享契约中定义输入、输出、错误和版本。
- 所有平台模块必须拥有独立的 `domain`、`service`、`handler`、`repository` 边界。
- 所有固件模块必须拥有公开头文件、私有实现、测试、Kconfig 和构建条件。
- 任何可选模块关闭后，其他已启用模块必须能够编译、启动和通过测试。
- 发行配置只决定模块是否打包，不改变功能模块的实现和测试。

固件模块开关统一使用：

```text
CONFIG_FEATURE_<MODULE_NAME>
```

例如：

```text
CONFIG_FEATURE_AUDIO_INPUT
CONFIG_FEATURE_CAMERA_VISION
CONFIG_FEATURE_DISPLAY_UI
CONFIG_FEATURE_CELLULAR_4G
CONFIG_FEATURE_MOTION_CONTROL
```
