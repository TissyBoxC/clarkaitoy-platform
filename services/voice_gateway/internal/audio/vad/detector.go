// Package vad detects voice activity.
package vad

// Detector reports whether an audio frame contains speech.
type Detector interface {
	IsSpeech(data []byte) bool
}
