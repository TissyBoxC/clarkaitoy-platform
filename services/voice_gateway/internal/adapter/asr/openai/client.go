// Package openai contains the OpenAI ASR adapter.
package openai

import (
	"context"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/adapter/asr"
)

// Client implements the ASR adapter using OpenAI.
type Client struct{}

// Open starts an OpenAI recognition stream.
func (c *Client) Open(_ context.Context, _ asr.Config) (asr.Stream, error) {
	return nil, nil
}
