// Package llm defines text conversation gateway contracts.
package llm

import "context"

// Message is one chat message.
type Message struct {
	Role    string
	Content string
}

// Request describes one chat request.
type Request struct {
	Model    string
	Messages []Message
}

// Stream emits model response chunks.
type Stream interface {
	Chunks() <-chan string
	Close() error
}

// Client connects the voice gateway to an LLM gateway.
type Client interface {
	Chat(ctx context.Context, request Request) (Stream, error)
}
