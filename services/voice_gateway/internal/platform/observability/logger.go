// Package observability owns logs, metrics, and traces.
package observability

// Logger is the service logging abstraction.
type Logger interface {
	Info(message string, fields ...any)
	Error(message string, fields ...any)
}
