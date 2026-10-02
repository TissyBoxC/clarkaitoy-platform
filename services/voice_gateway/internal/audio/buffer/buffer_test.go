package buffer

import (
	"errors"
	"testing"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/frame"
)

func newTestFrame(sequence uint32) frame.Frame {
	return frame.New(
		"audio_frame_0001",
		"device_0001",
		"stream_0001",
		sequence,
		time.Now().UTC(),
		[]byte{0x01},
	)
}

func TestPushPopPreservesOrder(t *testing.T) {
	ring := NewRingBuffer(4)
	for sequence := uint32(0); sequence < 3; sequence++ {
		if err := ring.Push(newTestFrame(sequence)); err != nil {
			t.Fatalf("push sequence %d: %v", sequence, err)
		}
	}

	for sequence := uint32(0); sequence < 3; sequence++ {
		popped, ok := ring.Pop()
		if !ok {
			t.Fatalf("expected frame %d", sequence)
		}
		if popped.Sequence != sequence {
			t.Fatalf("expected sequence %d, got %d", sequence, popped.Sequence)
		}
	}
	if _, ok := ring.Pop(); ok {
		t.Fatal("expected empty buffer")
	}
}

func TestPushRejectsDuplicateSequence(t *testing.T) {
	ring := NewRingBuffer(4)
	if err := ring.Push(newTestFrame(5)); err != nil {
		t.Fatalf("push first frame: %v", err)
	}
	if err := ring.Push(newTestFrame(5)); !errors.Is(err, ErrDuplicateFrame) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}

func TestOverflowEvictsOldestFrame(t *testing.T) {
	ring := NewRingBuffer(2)
	for sequence := uint32(0); sequence < 3; sequence++ {
		if err := ring.Push(newTestFrame(sequence)); err != nil {
			t.Fatalf("push sequence %d: %v", sequence, err)
		}
	}

	if ring.Len() != 2 {
		t.Fatalf("expected 2 buffered frames, got %d", ring.Len())
	}
	popped, _ := ring.Pop()
	if popped.Sequence != 1 {
		t.Fatalf("expected oldest surviving sequence 1, got %d", popped.Sequence)
	}
	if stats := ring.Stats(); stats.Dropped != 1 {
		t.Fatalf("expected 1 dropped frame, got %d", stats.Dropped)
	}
}

func TestSequenceGapIsCounted(t *testing.T) {
	ring := NewRingBuffer(4)
	if err := ring.Push(newTestFrame(1)); err != nil {
		t.Fatalf("push frame 1: %v", err)
	}
	if err := ring.Push(newTestFrame(4)); err != nil {
		t.Fatalf("push frame 4: %v", err)
	}
	if stats := ring.Stats(); stats.Gaps != 1 {
		t.Fatalf("expected 1 gap, got %d", stats.Gaps)
	}
}

func TestWrappedSequenceIsNotTreatedAsGap(t *testing.T) {
	ring := NewRingBuffer(4)
	if err := ring.Push(newTestFrame(^uint32(0))); err != nil {
		t.Fatalf("push max sequence: %v", err)
	}
	if err := ring.Push(newTestFrame(0)); err != nil {
		t.Fatalf("push wrapped sequence: %v", err)
	}
	if stats := ring.Stats(); stats.Gaps != 0 {
		t.Fatalf("expected no gap across the wrap, got %d", stats.Gaps)
	}
}

func TestResetClearsFrames(t *testing.T) {
	ring := NewRingBuffer(4)
	if err := ring.Push(newTestFrame(1)); err != nil {
		t.Fatalf("push frame: %v", err)
	}
	ring.Reset()
	if ring.Len() != 0 {
		t.Fatalf("expected empty buffer after reset, got %d", ring.Len())
	}
}

func TestInvalidCapacityFallsBackToDefault(t *testing.T) {
	ring := NewRingBuffer(MaxCapacityFrames + 1)
	if ring.capacity != DefaultCapacityFrames {
		t.Fatalf("expected default capacity, got %d", ring.capacity)
	}
}

func TestPushRejectsInvalidFrame(t *testing.T) {
	ring := NewRingBuffer(2)
	invalid := frame.Frame{Sequence: 1}
	if err := ring.Push(invalid); !errors.Is(err, ErrInvalidFrame) {
		t.Fatalf("expected invalid frame rejection, got %v", err)
	}
}
