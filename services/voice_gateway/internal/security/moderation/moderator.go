// Package moderation provides content moderation integration.
package moderation

// Moderator checks content for unsafe material.
type Moderator interface {
	Check(text string) (bool, error)
}
