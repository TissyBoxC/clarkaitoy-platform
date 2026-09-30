// Package local_whisper contains the local Whisper ASR adapter.
package local_whisper

import (
	"context"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/adapter/asr"
)

// Client implements the ASR adapter using a local Whisper service.
type Client struct{}

// Open starts a local Whisper recognition stream.
func (c *Client) Open(_ context.Context, _ asr.Config) (asr.Stream, error) {
	return nil, nil
}
