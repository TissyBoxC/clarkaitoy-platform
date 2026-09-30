// Package domain contains telemetry domain types.
package domain

// Heartbeat is a periodic device status message.
type Heartbeat struct {
	DeviceID string
	Firmware string
	Online   bool
}
