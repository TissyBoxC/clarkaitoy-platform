// Package domain contains audit domain types.
package domain

// Record describes one audited operation.
type Record struct {
	ActorID string
	Action  string
	Target  string
}
