// Package aliyun contains the Alibaba Cloud TTS adapter.
package aliyun

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/tts"
)

// Client implements the TTS adapter using Alibaba Cloud.
type Client struct{}

// Synthesize starts an Alibaba Cloud synthesis stream.
func (c *Client) Synthesize(_ context.Context, _ tts.Request) (tts.Stream, error) {
	return nil, nil
}
