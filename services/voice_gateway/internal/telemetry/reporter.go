// Package telemetry reports voice gateway health and latency.
package telemetry

// Reporter emits service telemetry.
type Reporter interface {
	Counter(name string, value int64)
	Histogram(name string, value float64)
}
