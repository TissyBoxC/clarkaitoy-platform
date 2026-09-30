// Package tencent contains the Tencent Cloud ASR adapter.
package tencent

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/asr"
)

// Client implements the ASR adapter using Tencent Cloud.
type Client struct{}

// Open starts a Tencent Cloud recognition stream.
func (c *Client) Open(_ context.Context, _ asr.Config) (asr.Stream, error) {
	return nil, nil
}
