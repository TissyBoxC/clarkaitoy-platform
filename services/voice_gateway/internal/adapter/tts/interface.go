// Package tts defines speech synthesis adapter contracts.
package tts

import "context"

// Request describes one synthesis request.
type Request struct {
	Text     string
	Voice    string
	Language string
}

// Stream produces synthesized audio.
type Stream interface {
	Audio() <-chan []byte
	Close() error
}

// Synthesizer creates synthesis streams.
type Synthesizer interface {
	Synthesize(ctx context.Context, request Request) (Stream, error)
}
