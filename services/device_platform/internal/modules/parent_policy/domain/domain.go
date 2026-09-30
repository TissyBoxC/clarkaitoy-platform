// Package domain contains parent policy domain types.
package domain

// Policy defines limits applied to a child and device.
type Policy struct {
	ChildID          string
	DailyMinutes     int
	AllowedContent   []string
	DisabledHours    []string
}
