// Package security owns authentication and authorization primitives.
package security

// TokenIssuer issues short-lived tokens for devices and clients.
type TokenIssuer interface {
	Issue(subject string) (string, error)
	Verify(token string) (string, error)
}
