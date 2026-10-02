package frame

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEncodeDecodeRoundTripPreservesFrame(t *testing.T) {
	capturedAt := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	original := New(
		"audio_frame_0001",
		"device_0001",
		"stream_0001",
		42,
		capturedAt,
		[]byte{0x4f, 0x70, 0x75, 0x73},
	)

	encoded, err := original.EncodeJSON()
	if err != nil {
		t.Fatalf("encode frame: %v", err)
	}

	decoded, err := DecodeJSON(encoded)
	if err != nil {
		t.Fatalf("decode frame: %v", err)
	}
	if decoded.FrameID != original.FrameID ||
		decoded.Sequence != original.Sequence ||
		string(decoded.Payload) != string(original.Payload) {
		t.Fatalf("round trip changed frame: %+v", decoded)
	}
}

func TestEncodeJSONMatchesWireContractFields(t *testing.T) {
	audioFrame := New(
		"audio_frame_0001",
		"device_0001",
		"stream_0001",
		7,
		time.Unix(0, 0).UTC(),
		[]byte{0x01},
	)
	encoded, err := audioFrame.EncodeJSON()
	if err != nil {
		t.Fatalf("encode frame: %v", err)
	}

	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if wire["schema_version"] != SchemaVersion {
		t.Fatalf("unexpected schema_version: %v", wire["schema_version"])
	}
	if wire["sample_rate_hz"].(float64) != SampleRateHz {
		t.Fatalf("unexpected sample_rate_hz: %v", wire["sample_rate_hz"])
	}
	if wire["channel_count"].(float64) != ChannelCount {
		t.Fatalf("unexpected channel_count: %v", wire["channel_count"])
	}
	if wire["duration_ms"].(float64) != DurationMS {
		t.Fatalf("unexpected duration_ms: %v", wire["duration_ms"])
	}
	if wire["encoding"] != EncodingName {
		t.Fatalf("unexpected encoding: %v", wire["encoding"])
	}
}

func TestDecodeJSONRejectsProfileMismatch(t *testing.T) {
	valid := `{
		"schema_version": "1.0.0",
		"frame_id": "audio_frame_0001",
		"device_id": "device_0001",
		"stream_id": "stream_0001",
		"sequence": 1,
		"captured_at": "2026-10-03T10:00:00Z",
		"sample_rate_hz": 48000,
		"channel_count": 1,
		"duration_ms": 20,
		"encoding": "opus",
		"payload_base64": "AQ=="
	}`

	_, err := DecodeJSON([]byte(valid))
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("expected profile mismatch, got %v", err)
	}
}

func TestValidateRejectsOversizedPayload(t *testing.T) {
	audioFrame := New(
		"audio_frame_0001",
		"device_0001",
		"stream_0001",
		1,
		time.Now().UTC(),
		make([]byte, MaxPayloadBytes+1),
	)
	if err := audioFrame.Validate(); !errors.Is(err, ErrPayloadTooLarge) {
		t.Fatalf("expected oversize rejection, got %v", err)
	}
}

func TestValidateRejectsEmptyIdentifier(t *testing.T) {
	audioFrame := New(
		"",
		"device_0001",
		"stream_0001",
		1,
		time.Now().UTC(),
		[]byte{0x01},
	)
	if err := audioFrame.Validate(); !errors.Is(err, ErrMissingIdentifier) {
		t.Fatalf("expected missing identifier rejection, got %v", err)
	}
}

func TestDecodeJSONRejectsInvalidBase64(t *testing.T) {
	document := `{
		"schema_version": "1.0.0",
		"frame_id": "audio_frame_0001",
		"device_id": "device_0001",
		"stream_id": "stream_0001",
		"sequence": 1,
		"captured_at": "2026-10-03T10:00:00Z",
		"sample_rate_hz": 16000,
		"channel_count": 1,
		"duration_ms": 20,
		"encoding": "opus",
		"payload_base64": "not base64!!"
	}`
	_, err := DecodeJSON([]byte(document))
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected invalid payload rejection, got %v", err)
	}
}

func TestDecodeJSONRejectsUnknownSchemaVersion(t *testing.T) {
	document := `{
		"schema_version": "2.0.0",
		"frame_id": "audio_frame_0001",
		"device_id": "device_0001",
		"stream_id": "stream_0001",
		"sequence": 1,
		"captured_at": "2026-10-03T10:00:00Z",
		"sample_rate_hz": 16000,
		"channel_count": 1,
		"duration_ms": 20,
		"encoding": "opus",
		"payload_base64": "AQ=="
	}`
	_, err := DecodeJSON([]byte(document))
	if !errors.Is(err, ErrInvalidSchemaVersion) {
		t.Fatalf("expected schema version rejection, got %v", err)
	}
	if !strings.Contains(err.Error(), "2.0.0") {
		t.Fatalf("error should identify the rejected version: %v", err)
	}
}
