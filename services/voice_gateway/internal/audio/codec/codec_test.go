package codec

import (
	"errors"
	"math"
	"testing"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/frame"
)

func TestEncodeDecodeRoundTripProducesRecognizablePCM(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}

	// A 440 Hz tone at a quarter of full scale is a stable, non-silent
	// fixture that does not depend on random data.
	pcm := make([]int16, frame.SamplesPerFrame*frame.ChannelCount)
	for index := range pcm {
		phase := 2 * math.Pi * 440 * float64(index) / float64(frame.SampleRateHz)
		pcm[index] = int16(8192 * math.Sin(phase))
	}

	encoded := make([]byte, frame.MaxPayloadBytes)
	encodedBytes, err := audioCodec.EncodePCM(pcm, encoded)
	if err != nil {
		t.Fatalf("encode pcm: %v", err)
	}
	if encodedBytes == 0 || encodedBytes > frame.MaxPayloadBytes {
		t.Fatalf("unexpected encoded size: %d", encodedBytes)
	}

	decoded := make([]int16, frame.SamplesPerFrame*frame.ChannelCount)
	samples, err := audioCodec.DecodePacket(encoded[:encodedBytes], decoded)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if samples != frame.SamplesPerFrame {
		t.Fatalf("expected %d samples, got %d", frame.SamplesPerFrame, samples)
	}

	var peak int16
	for _, sample := range decoded {
		if sample > peak {
			peak = sample
		}
	}
	if peak == 0 {
		t.Fatal("decoded audio is silent")
	}
}

func TestEncodePCMRejectsWrongFrameSize(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}

	_, err = audioCodec.EncodePCM(make([]int16, 10), make([]byte, frame.MaxPayloadBytes))
	if !errors.Is(err, ErrInvalidPCMWindows) {
		t.Fatalf("expected frame size rejection, got %v", err)
	}
}

func TestDecodePacketRejectsOversizedPacket(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}

	_, err = audioCodec.DecodePacket(
		make([]byte, frame.MaxPayloadBytes+1),
		make([]int16, frame.SamplesPerFrame),
	)
	if !errors.Is(err, ErrInvalidPacketSize) {
		t.Fatalf("expected packet size rejection, got %v", err)
	}
}

func TestDecodePacketAcceptsEmptyPacketAsConcealment(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}

	decoded := make([]int16, frame.SamplesPerFrame)
	samples, err := audioCodec.DecodePacket(nil, decoded)
	if err != nil {
		t.Fatalf("decode concealment: %v", err)
	}
	if samples != frame.SamplesPerFrame {
		t.Fatalf("expected concealed frame of %d samples, got %d", frame.SamplesPerFrame, samples)
	}
}

func TestResetRestoresUsableCodec(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}
	if err := audioCodec.Reset(); err != nil {
		t.Fatalf("reset codec: %v", err)
	}

	pcm := make([]int16, frame.SamplesPerFrame)
	encoded := make([]byte, frame.MaxPayloadBytes)
	if _, err := audioCodec.EncodePCM(pcm, encoded); err != nil {
		t.Fatalf("encode after reset: %v", err)
	}
}

func TestDecodePacketRequiresFullFrameBuffer(t *testing.T) {
	audioCodec, err := NewOpusCodec()
	if err != nil {
		t.Fatalf("create codec: %v", err)
	}

	_, err = audioCodec.DecodePacket([]byte{0x01}, make([]int16, 8))
	if !errors.Is(err, ErrInvalidPCMWindows) {
		t.Fatalf("expected pcm window rejection, got %v", err)
	}
}
