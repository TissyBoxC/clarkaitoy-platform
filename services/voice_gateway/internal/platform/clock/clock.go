// Package clock provides an injectable time source.
package clock

import "time"

// Clock abstracts the current time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// SystemClock uses the host clock.
type SystemClock struct{}

// Now returns the current local time.
func (SystemClock) Now() time.Time {
	return time.Now()
}
