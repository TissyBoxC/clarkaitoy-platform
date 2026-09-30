// Package domain contains child profile domain types.
package domain

// Child represents a child profile owned by a family.
type Child struct {
	ID       string
	FamilyID string
	Name     string
	Age      int
}
