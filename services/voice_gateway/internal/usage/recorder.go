// Package usage records AI and voice usage.
package usage

import (
	"context"
	"time"
)

// Record describes one billable or observable usage event.
type Record struct {
	DeviceID   string
	Model      string
	Latency    time.Duration
	InputSize  int
	OutputSize int
	SpentUSD   float64
}

// Recorder stores usage events.
type Recorder interface {
	Record(ctx context.Context, record Record) error
}
