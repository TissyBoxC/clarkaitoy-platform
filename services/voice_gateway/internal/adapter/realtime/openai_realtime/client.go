// Package openai_realtime contains the OpenAI Realtime adapter.
package openai_realtime

import (
	"context"

	"github.com/clarkaitoy/voice_gateway/internal/adapter/realtime"
)

// Client connects to the OpenAI Realtime API.
type Client struct{}

// Connect opens an OpenAI Realtime session.
func (c *Client) Connect(_ context.Context) (realtime.Session, error) {
	return nil, nil
}
