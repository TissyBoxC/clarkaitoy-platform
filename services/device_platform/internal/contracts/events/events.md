# Device Platform Events

Reserved event contracts for asynchronous processing.

```text
device.bound
device.unbound
device.online
device.offline
policy.updated
content.released
ota.started
ota.completed
ota.failed
```

Events must be idempotent and carry a stable event ID.
