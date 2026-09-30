// Package local_piper contains the local Piper TTS adapter.
package local_piper

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/tts"
)

// Client implements the TTS adapter using a local Piper service.
type Client struct{}

// Synthesize starts a local Piper synthesis stream.
func (c *Client) Synthesize(_ context.Context, _ tts.Request) (tts.Stream, error) {
	return nil, nil
}
