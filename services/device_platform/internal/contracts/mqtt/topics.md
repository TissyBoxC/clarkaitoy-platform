# MQTT Topics

Reserved topic layout for the device platform.

```text
clarkaitoy/v1/devices/{device_id}/heartbeat
clarkaitoy/v1/devices/{device_id}/status
clarkaitoy/v1/devices/{device_id}/policy
clarkaitoy/v1/devices/{device_id}/ota
```

All payloads must include `schema_version`, `device_id`, and `sent_at`.
