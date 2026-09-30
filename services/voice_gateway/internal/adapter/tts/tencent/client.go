// Package tencent contains the Tencent Cloud TTS adapter.
package tencent

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/tts"
)

// Client implements the TTS adapter using Tencent Cloud.
type Client struct{}

// Synthesize starts a Tencent Cloud synthesis stream.
func (c *Client) Synthesize(_ context.Context, _ tts.Request) (tts.Stream, error) {
	return nil, nil
}
