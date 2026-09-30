// Package domain contains OTA domain types.
package domain

// Release describes a firmware release.
type Release struct {
	Version string
	Channel string
	URL     string
}
