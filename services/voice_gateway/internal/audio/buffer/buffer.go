// Package buffer provides bounded, sequence-aware buffering for inbound realtime
// audio.
//
// The gateway cannot assume packets arrive in order or without loss, and it must
// not let one slow stream grow without bound. The ring buffer keeps a fixed
// window of the newest frames, reports gaps, and rejects duplicates so a
// reconnecting device cannot replay stale audio into an active conversation.
package buffer

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/audio/frame"
)

const (
	// DefaultCapacityFrames holds 400 ms of audio at the 20 ms frame profile.
	DefaultCapacityFrames = 20

	// MaxCapacityFrames bounds memory per stream. One second is enough to
	// absorb jitter without holding raw child speech longer than necessary.
	MaxCapacityFrames = 50
)

// Errors returned by ring buffer operations.
var (
	ErrBufferClosed   = errors.New("audio buffer is closed")
	ErrInvalidFrame   = errors.New("audio frame is invalid")
	ErrDuplicateFrame = errors.New("audio frame sequence already buffered")
)

// RingBuffer stores the newest audio frames for one stream.
//
// The zero value is not usable; call NewRingBuffer. Methods are safe for
// concurrent use because the WebSocket reader and the ASR consumer run on
// different goroutines.
type RingBuffer struct {
	mutex       sync.Mutex
	frames      []frame.Frame
	capacity    int
	next        int
	size        int
	hasBaseline bool
	lastSeq     uint32
	dropped     uint64
	gaps        uint64
}

// NewRingBuffer returns a bounded buffer. Capacity outside
// [1, MaxCapacityFrames] is clamped to DefaultCapacityFrames.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 1 || capacity > MaxCapacityFrames {
		capacity = DefaultCapacityFrames
	}
	return &RingBuffer{
		frames:   make([]frame.Frame, capacity),
		capacity: capacity,
	}
}

// Push appends one decoded frame in sequence order.
//
// A duplicate sequence returns ErrDuplicateFrame. A sequence jump increments
// the gap counter so telemetry can distinguish loss from network delay. The
// oldest frame is evicted when the window overflows.
func (b *RingBuffer) Push(audioFrame frame.Frame) error {
	if b == nil {
		return ErrInvalidFrame
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.capacity == 0 {
		return ErrBufferClosed
	}
	if err := audioFrame.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFrame, err)
	}

	if b.hasBaseline {
		if audioFrame.Sequence == b.lastSeq {
			return ErrDuplicateFrame
		}
		if isNewerSequence(audioFrame.Sequence, b.lastSeq) {
			// Sequence numbers wrap at 32 bits; a wrapped difference of more
			// than one frame means at least one packet never arrived.
			if audioFrame.Sequence-b.lastSeq > 1 {
				b.gaps++
			}
		}
	}

	if b.size == b.capacity {
		b.frames[b.next] = frame.Frame{}
		b.dropped++
	} else {
		b.size++
	}
	b.frames[b.next] = audioFrame
	b.next = (b.next + 1) % b.capacity
	b.hasBaseline = true
	b.lastSeq = audioFrame.Sequence
	return nil
}

// Pop removes and returns the oldest buffered frame.
//
// The second return value is false when the buffer is empty.
func (b *RingBuffer) Pop() (frame.Frame, bool) {
	if b == nil {
		return frame.Frame{}, false
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.size == 0 {
		return frame.Frame{}, false
	}
	start := (b.next - b.size + b.capacity) % b.capacity
	audioFrame := b.frames[start]
	b.frames[start] = frame.Frame{}
	b.size--
	return audioFrame, true
}

// Len returns the number of buffered frames.
func (b *RingBuffer) Len() int {
	if b == nil {
		return 0
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.size
}

// Stats reports bounded counters for telemetry and diagnostics.
func (b *RingBuffer) Stats() Stats {
	if b == nil {
		return Stats{}
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return Stats{
		Buffered: b.size,
		Dropped:  b.dropped,
		Gaps:     b.gaps,
	}
}

// Reset discards buffered audio and sequence state between conversations.
func (b *RingBuffer) Reset() {
	if b == nil {
		return
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for index := range b.frames {
		b.frames[index] = frame.Frame{}
	}
	b.next = 0
	b.size = 0
	b.hasBaseline = false
	b.lastSeq = 0
}

// Stats is a point-in-time view of one buffer.
type Stats struct {
	Buffered int
	Dropped  uint64
	Gaps     uint64
}

// isNewerSequence reports whether candidate is ahead of current using signed
// 32-bit comparison so wraparound stays monotonic.
func isNewerSequence(candidate uint32, current uint32) bool {
	return int32(candidate-current) > 0
}
