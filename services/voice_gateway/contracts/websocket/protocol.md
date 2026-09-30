# Voice WebSocket Protocol

Initial protocol reserved for device audio sessions.

## Connection

```text
GET /v1/voice
Authorization: Bearer <short-lived-device-token>
```

## Client Messages

```text
audio.start
audio.frame
audio.end
audio.cancel
session.close
```

## Server Messages

```text
session.ready
asr.partial
asr.final
llm.delta
tts.audio
session.error
session.closed
```

Every control message includes `schema_version`, `session_id`, and `sequence`.
