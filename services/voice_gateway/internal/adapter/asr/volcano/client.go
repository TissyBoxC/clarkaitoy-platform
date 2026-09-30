// Package volcano contains the Volcano ASR adapter.
package volcano

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/asr"
)

// Client implements the ASR adapter using Volcano.
type Client struct{}

// Open starts a Volcano recognition stream.
func (c *Client) Open(_ context.Context, _ asr.Config) (asr.Stream, error) {
	return nil, nil
}
