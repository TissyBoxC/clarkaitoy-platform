// Package asr defines speech recognition adapter contracts.
package asr

import "context"

// Config controls one recognition stream.
type Config struct {
	Language string
}

// Result is one recognition result.
type Result struct {
	Text      string
	IsFinal   bool
	RequestID string
}

// Stream accepts audio and emits recognition results.
type Stream interface {
	WriteAudio(data []byte) error
	Results() <-chan Result
	Close() error
}

// Recognizer creates recognition streams.
type Recognizer interface {
	Open(ctx context.Context, config Config) (Stream, error)
}
