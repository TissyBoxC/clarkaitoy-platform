// Package codec converts realtime audio between the fixed PCM profile and the
// Opus packets carried by the device WebSocket.
//
// The gateway listens to 16 kHz mono Opus and hands PCM to the ASR adapter, so
// inbound packets are decoded to int16 PCM and outbound TTS PCM is encoded back
// to the same profile. The Opus implementation is pure Go so the service keeps
// building with CGO_ENABLED=0.
package codec

import (
	"errors"
	"fmt"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/frame"
	"github.com/thesyncim/gopus"
)

// Errors returned while converting audio. They are stable so the session layer
// can map them to session error codes without parsing messages.
var (
	ErrInvalidPacketSize = errors.New("audio packet size is invalid")
	ErrInvalidPCMWindows = errors.New("audio pcm buffer is smaller than one frame")
	ErrCodecUnavailable  = errors.New("audio codec is not initialized")
)

// Codec converts one encoded packet profile to and from PCM.
//
// Implementations are stateful: Opus keeps per-stream prediction state, so a
// codec must be used by exactly one stream and must not be shared between
// goroutines.
type Codec interface {
	// DecodePacket decodes one Opus packet into int16 PCM samples. It returns
	// the number of written samples across all channels.
	DecodePacket(packet []byte, pcm []int16) (int, error)

	// EncodePCM encodes exactly one frame of int16 PCM into destination. It
	// returns the number of encoded bytes.
	EncodePCM(pcm []int16, destination []byte) (int, error)

	// Reset clears per-stream codec state between conversations.
	Reset() error
}

// OpusCodec implements Codec for the fixed 16 kHz mono Opus frame profile.
type OpusCodec struct {
	encoder *gopus.Encoder
	decoder *gopus.Decoder
}

// NewOpusCodec builds a codec for the contract frame profile.
//
// The bitrate targets wideband child speech and the complexity stays moderate
// so a single gateway instance can serve many concurrent streams.
func NewOpusCodec() (*OpusCodec, error) {
	encoder, err := gopus.NewEncoder(gopus.EncoderConfig{
		SampleRate:  frame.SampleRateHz,
		Channels:    frame.ChannelCount,
		Application: gopus.ApplicationVoIP,
	})
	if err != nil {
		return nil, fmt.Errorf("create opus encoder: %w", err)
	}
	if err := encoder.SetBitrate(24000); err != nil {
		return nil, fmt.Errorf("configure opus bitrate: %w", err)
	}
	if err := encoder.SetComplexity(5); err != nil {
		return nil, fmt.Errorf("configure opus complexity: %w", err)
	}

	decoder, err := gopus.NewDecoder(gopus.DefaultDecoderConfig(
		frame.SampleRateHz,
		frame.ChannelCount,
	))
	if err != nil {
		return nil, fmt.Errorf("create opus decoder: %w", err)
	}

	return &OpusCodec{encoder: encoder, decoder: decoder}, nil
}

// DecodePacket decodes one Opus packet into PCM.
//
// An empty packet is treated as packet loss by Opus and produces concealed
// audio, which keeps a dropped frame from breaking the stream timeline.
func (c *OpusCodec) DecodePacket(packet []byte, pcm []int16) (int, error) {
	if c == nil || c.decoder == nil {
		return 0, ErrCodecUnavailable
	}
	if len(packet) > frame.MaxPayloadBytes {
		return 0, fmt.Errorf("%w: %d bytes", ErrInvalidPacketSize, len(packet))
	}
	if len(pcm) < frame.SamplesPerFrame*frame.ChannelCount {
		return 0, fmt.Errorf(
			"%w: %d samples",
			ErrInvalidPCMWindows,
			len(pcm),
		)
	}
	samples, err := c.decoder.DecodeInt16(packet, pcm)
	if err != nil {
		return 0, fmt.Errorf("decode opus packet: %w", err)
	}
	return samples, nil
}

// EncodePCM encodes exactly one frame of PCM into destination.
func (c *OpusCodec) EncodePCM(pcm []int16, destination []byte) (int, error) {
	if c == nil || c.encoder == nil {
		return 0, ErrCodecUnavailable
	}
	if len(pcm) != frame.SamplesPerFrame*frame.ChannelCount {
		return 0, fmt.Errorf("%w: %d samples", ErrInvalidPCMWindows, len(pcm))
	}
	if len(destination) == 0 {
		return 0, ErrInvalidPacketSize
	}
	encoded, err := c.encoder.EncodeInt16(pcm, destination)
	if err != nil {
		return 0, fmt.Errorf("encode opus packet: %w", err)
	}
	return encoded, nil
}

// Reset rebuilds codec state so a new conversation never inherits the previous
// stream's prediction history.
func (c *OpusCodec) Reset() error {
	rebuilt, err := NewOpusCodec()
	if err != nil {
		return err
	}
	*c = *rebuilt
	return nil
}
