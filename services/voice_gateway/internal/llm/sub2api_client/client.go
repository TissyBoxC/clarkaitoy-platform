// Package sub2api_client implements the OpenAI-compatible sub2api client.
package sub2api_client

import (
	"context"

	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/llm"
)

// Client sends chat requests to the sub2api AI gateway.
type Client struct {
	BaseURL string
	APIKey  string
}

// Chat starts a streaming chat request.
func (c *Client) Chat(_ context.Context, _ llm.Request) (llm.Stream, error) {
	return nil, nil
}
