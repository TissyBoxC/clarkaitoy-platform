// Package openai contains the OpenAI TTS adapter.
package openai

import (
	"context"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/adapter/tts"
)

// Client implements the TTS adapter using OpenAI.
type Client struct{}

// Synthesize starts an OpenAI synthesis stream.
func (c *Client) Synthesize(_ context.Context, _ tts.Request) (tts.Stream, error) {
	return nil, nil
}
