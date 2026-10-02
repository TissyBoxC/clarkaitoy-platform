// Package frame defines the realtime audio frame profile shared by the device
// firmware, the WebSocket transport, and the codec/buffer/playback layers.
//
// The profile is fixed at 16 kHz mono Opus with 20 ms frames. Both endpoints
// must agree on this shape so sequence numbers, timestamps, and buffer windows
// can be compared without negotiation.
package frame

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// SchemaVersion identifies the audio frame contract implemented here.
	SchemaVersion = "1.0.0"

	// SampleRateHz is fixed at 16 kHz for wideband child speech.
	SampleRateHz = 16000

	// ChannelCount is mono; the hardware profile has a single microphone.
	ChannelCount = 1

	// DurationMS is the Opus frame duration in milliseconds.
	DurationMS = 20

	// EncodingName is the canonical codec identifier of the contract.
	EncodingName = "opus"

	// SamplesPerFrame is the decoded sample count per channel for one frame.
	SamplesPerFrame = SampleRateHz * DurationMS / 1000

	// MaxPayloadBytes bounds one base64-encoded Opus packet. A 20 ms Opus packet
	// at the highest supported bitrate stays far below this limit, so a larger
	// value indicates a malformed or hostile frame.
	MaxPayloadBytes = 1024
)

// Errors returned while validating a frame. They are stable so callers can map
// them to device audio error codes without matching on message text.
var (
	ErrInvalidSchemaVersion = errors.New("audio frame schema_version is not supported")
	ErrMissingIdentifier    = errors.New("audio frame identifier is empty")
	ErrInvalidProfile       = errors.New("audio frame profile is not 16 kHz mono opus 20ms")
	ErrInvalidSequence      = errors.New("audio frame sequence is out of range")
	ErrInvalidTimestamp     = errors.New("audio frame captured_at is empty")
	ErrPayloadTooLarge      = errors.New("audio frame payload exceeds the frame limit")
	ErrInvalidPayload       = errors.New("audio frame payload is not valid base64")
)

// Frame is one timestamped Opus payload in the fixed realtime profile.
type Frame struct {
	SchemaVersion string
	FrameID       string
	DeviceID      string
	StreamID      string
	Sequence      uint32
	CapturedAt    time.Time
	Payload       []byte
}

// New creates a frame in the fixed profile. Payload ownership transfers to the
// returned frame; callers must not mutate the slice afterwards.
func New(
	frameID string,
	deviceID string,
	streamID string,
	sequence uint32,
	capturedAt time.Time,
	payload []byte,
) Frame {
	return Frame{
		SchemaVersion: SchemaVersion,
		FrameID:       frameID,
		DeviceID:      deviceID,
		StreamID:      streamID,
		Sequence:      sequence,
		CapturedAt:    capturedAt,
		Payload:       payload,
	}
}

// Validate reports whether the frame satisfies the audio frame contract.
//
// It rejects unknown schema versions, empty identifiers, out-of-range
// payloads, and non-monotonic timestamps so a malformed frame never reaches
// the codec or the ASR adapter.
func (f Frame) Validate() error {
	if f.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: %q", ErrInvalidSchemaVersion, f.SchemaVersion)
	}
	if f.FrameID == "" || f.DeviceID == "" || f.StreamID == "" {
		return ErrMissingIdentifier
	}
	if f.CapturedAt.IsZero() {
		return ErrInvalidTimestamp
	}
	if len(f.Payload) == 0 || len(f.Payload) > MaxPayloadBytes {
		return fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(f.Payload))
	}
	return nil
}

// WireFrame is the JSON representation defined by
// packages/contracts/schemas/audio_frame.schema.json.
type WireFrame struct {
	SchemaVersion string `json:"schema_version"`
	FrameID       string `json:"frame_id"`
	DeviceID      string `json:"device_id"`
	StreamID      string `json:"stream_id"`
	Sequence      uint32 `json:"sequence"`
	CapturedAt    string `json:"captured_at"`
	SampleRateHz  int    `json:"sample_rate_hz"`
	ChannelCount  int    `json:"channel_count"`
	DurationMS    int    `json:"duration_ms"`
	Encoding      string `json:"encoding"`
	PayloadBase64 string `json:"payload_base64"`
}

// EncodeJSON serializes the frame into the wire contract.
func (f Frame) EncodeJSON() ([]byte, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	wire := WireFrame{
		SchemaVersion: SchemaVersion,
		FrameID:       f.FrameID,
		DeviceID:      f.DeviceID,
		StreamID:      f.StreamID,
		Sequence:      f.Sequence,
		CapturedAt:    f.CapturedAt.UTC().Format(time.RFC3339Nano),
		SampleRateHz:  SampleRateHz,
		ChannelCount:  ChannelCount,
		DurationMS:    DurationMS,
		Encoding:      EncodingName,
		PayloadBase64: base64.StdEncoding.EncodeToString(f.Payload),
	}
	return json.Marshal(wire)
}

// DecodeJSON parses a wire frame and enforces the fixed profile. A frame that
// declares a different sample rate, channel count, duration, or codec is
// rejected instead of being silently reinterpreted.
func DecodeJSON(data []byte) (Frame, error) {
	var wire WireFrame
	if err := json.Unmarshal(data, &wire); err != nil {
		return Frame{}, fmt.Errorf("decode audio frame: %w", err)
	}
	if wire.SchemaVersion != SchemaVersion {
		return Frame{}, fmt.Errorf("%w: %q", ErrInvalidSchemaVersion, wire.SchemaVersion)
	}
	if wire.SampleRateHz != SampleRateHz ||
		wire.ChannelCount != ChannelCount ||
		wire.DurationMS != DurationMS ||
		wire.Encoding != EncodingName {
		return Frame{}, ErrInvalidProfile
	}
	payload, err := base64.StdEncoding.DecodeString(wire.PayloadBase64)
	if err != nil {
		return Frame{}, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	capturedAt, err := time.Parse(time.RFC3339Nano, wire.CapturedAt)
	if err != nil {
		return Frame{}, fmt.Errorf("%w: %v", ErrInvalidTimestamp, err)
	}
	frame := Frame{
		SchemaVersion: wire.SchemaVersion,
		FrameID:       wire.FrameID,
		DeviceID:      wire.DeviceID,
		StreamID:      wire.StreamID,
		Sequence:      wire.Sequence,
		CapturedAt:    capturedAt,
		Payload:       payload,
	}
	if err := frame.Validate(); err != nil {
		return Frame{}, err
	}
	return frame, nil
}
