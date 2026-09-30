// Package realtime defines end-to-end realtime voice adapter contracts.
package realtime

import "context"

// Session is one bidirectional realtime voice session.
type Session interface {
	WriteAudio([]byte) error
	Audio() <-chan []byte
	Close() error
}

// Connector opens realtime voice sessions.
type Connector interface {
	Connect(ctx context.Context) (Session, error)
}
