// Package usage records AI and voice usage.
package usage

import "time"

// Record describes one billable or observable usage event.
type Record struct {
	DeviceID   string
	Model      string
	Latency    time.Duration
	InputSize  int
	OutputSize int
}

// Recorder stores usage events.
type Recorder interface {
	Record(record Record) error
}
