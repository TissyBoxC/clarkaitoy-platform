// Package frame defines realtime audio frame types.
package frame

import "time"

// AudioFrame is one timestamped audio payload.
type AudioFrame struct {
	Data      []byte
	Timestamp time.Time
}
