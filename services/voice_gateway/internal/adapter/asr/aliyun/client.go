// Package aliyun contains the Alibaba Cloud ASR adapter.
package aliyun

import (
	"context"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/adapter/asr"
)

// Client implements the ASR adapter using Alibaba Cloud.
type Client struct{}

// Open starts an Alibaba Cloud recognition stream.
func (c *Client) Open(_ context.Context, _ asr.Config) (asr.Stream, error) {
	return nil, nil
}
