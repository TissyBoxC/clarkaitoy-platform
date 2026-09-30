// Package domain contains content domain types.
package domain

// ContentPackage groups downloadable child content.
type ContentPackage struct {
	ID       string
	Title    string
	AgeRange string
	Version  string
}
