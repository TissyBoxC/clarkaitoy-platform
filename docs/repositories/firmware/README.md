# 固件仓库代码文档

## 1. 仓库信息

| 项目 | 内容 |
| --- | --- |
| 仓库 | `clarkaitoy-firmware` |
| 本地路径 | `D:\service\clarkaitoy\firmware` |
| 默认分支 | `main` |
| 远端 | `https://github.com/TissyBoxC/clarkaitoy-firmware.git` |
| 目标硬件 | ESP32-S3 N16R8 |
| 资源 | 16 MB Flash、8 MB OPI PSRAM |
| 框架 | ESP-IDF |
| 构建工具 | PlatformIO |
| 构建目录 | `bgen/` |

固件负责设备端能力，不保存儿童长期档案、模型供应商密钥、内容运营数据或家长
账号。所有功能都必须独立模块化，构建时可以选择包含或排除任意模块。

## 2. 目标目录

```text
firmware/
  src/
    main.c                         仅组合和启动模块
  components/
    system_core/
    module_registry/
    version_info/
    error_recovery/
    config_store/
    audio_input/
    audio_output/
    audio_pipeline/
    playback_queue/
    voice_wake/
    wake_feedback/
    voice_session/
    conversation_context/
    child_prompt_profile/
    asr_client/
    llm_client/
    tts_client/
    content_library/
    content_story/
    content_nursery_rhyme/
    content_poetry/
    content_english/
    content_encyclopedia/
    content_bedtime/
    content_age_tier/
    parent_link/
    parent_policy/
    usage_report/
    network_manager/
    wifi_provisioning/
    time_sync/
    cloud_auth/
    privacy_guard/
    content_filter/
    transport_security/
    ota_manager/
    ota_download/
    ota_validate/
    ota_rollback/
    device_interaction/
    button_input/
    led_indicator/
    prompt_tone/
    volume_control/
    factory_reset/
    full_duplex_voice/
    far_field_audio/
    continuous_conversation/
    content_package_manager/
    theme_package/
    content_update/
    favorites/
    playback_history/
    companionship/
    emotion_response/
    encouragement/
    growth_record/
    habit_reminder/
    alarm_reminder/
    long_term_memory/
    english_learning/
    word_practice/
    sentence_practice/
    speaking_practice/
    pronunciation_feedback/
    remote_message/
    remote_playback/
    device_status/
    content_recommendation/
    offline_fallback/
    offline_story/
    local_commands/
    cache_playback/
    system_diagnostics/
    log_service/
    crash_report/
    network_quality/
    temperature_monitor/
    camera_vision/
    camera_capture/
    object_recognition/
    picture_book_recognition/
    photo_qa/
    image_privacy/
    display_ui/
    expression_animation/
    point_read/
    video_playback/
    eye_care_display/
    learning_visualization/
    touch_input/
    gesture_input/
    haptic_feedback/
    cellular_4g/
    esim_management/
    data_usage_management/
    remote_wake/
    battery_power/
    charge_management/
    battery_estimation/
    low_power/
    deep_sleep/
    motion_control/
    motor_drive/
    servo_drive/
    walk_action/
    dance_action/
    tail_action/
    form_factor_profile/
    video_call/
    remote_companionship/
    video_session/
    wearable_location/
    phone_call/
    geofence/
    sos/
    family_linkage/
    device_mesh/
    content_sync/
    bluetooth_audio/
  boards/
  partitions.csv
  platformio.ini
  sdkconfig.defaults
  test/
```

## 3. 模块通用结构

每个模块必须使用相同结构：

```text
components/<module_name>/
  CMakeLists.txt
  Kconfig
  include/
    <module_name>.h
  src/
    <module_name>.c
  test/
    <module_name>_test.c
  README.md
```

`CMakeLists.txt` 只有在 `CONFIG_FEATURE_<MODULE_NAME>` 打开时才编译源码和链接
依赖。`main.c` 通过 `module_registry` 注册模块，不直接包含业务实现。

## 4. 模块依赖规则

- `src/main.c` 只负责初始化、注册和启动已启用模块。
- 模块只能包含其他模块的公开头文件。
- 依赖必须单向，禁止循环依赖。
- 业务模块不能直接操作 GPIO、I2C、SPI、LCD 或摄像头寄存器。
- 硬件访问封装在 HAL、driver 或 adapter 模块中。
- 可选模块必须有独立 Kconfig 开关。
- 删除模块时不修改无关模块，只更新组合根、构建配置和依赖声明。
- 模块测试在 `pio test` 中独立运行。

## 5. 功能覆盖

固件必须覆盖
[功能覆盖矩阵](../../features/feature-coverage.md)中的所有基础层、主流层和
差异化层功能。完整模块清单、所属层级、优先级和依赖见
[固件模块目录](module-catalog.md)。

## 6. 构建与发行

构建、开关组合和版本发布见
[固件构建与发布](build-and-release.md)。所有构建物、镜像、中间文件和测试结果
必须输出到 `bgen/`，该目录不进入 Git。

## 7. 最佳实践

- 上层业务通过公开接口调用驱动，不直接依赖具体硬件。
- 断电、重启、网络异常和模块失败必须可恢复或可诊断。
- 设备状态、策略版本、固件版本和 capability 必须可上报。
- 音频、图片、位置和儿童相关数据必须执行最小化采集和隐私授权。
- 所有可选模块关闭后，基础音频版仍能独立构建和运行。
