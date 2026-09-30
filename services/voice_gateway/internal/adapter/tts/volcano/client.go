// Package volcano contains the Volcano TTS adapter.
package volcano

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/tts"
)

// Client implements the TTS adapter using Volcano.
type Client struct{}

// Synthesize starts a Volcano synthesis stream.
func (c *Client) Synthesize(_ context.Context, _ tts.Request) (tts.Stream, error) {
	return nil, nil
}
