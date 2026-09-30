// Package domain contains remote UI text domain types.
package domain

// UITextPackage is a signed, immutable version of user-visible text for one
// platform and language.
type UITextPackage struct {
	ID               string
	PackageVersion   int
	Language         string
	Platform         string
	MinClientVersion string
	Status           string
	PublishedAt      string
}

// UITextEntry is one user-visible value addressed by a stable key.
type UITextEntry struct {
	UITextKey string
	Value     string
	MaxLength int
	Variables []string
}
