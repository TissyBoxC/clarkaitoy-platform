// Package domain contains AI gateway account and credential types.
package domain

import (
	"errors"
	"time"
)

var (
	ErrAccountNotFound     = errors.New("AI account not found")
	ErrProviderRejected    = errors.New("AI provider rejected request")
	ErrProviderUnavailable = errors.New("AI provider unavailable")
	ErrCredentialInvalid   = errors.New("AI credential is invalid")
)

// Account mirrors the platform's view of one provider execution account.
type Account struct {
	ID                   string
	ParentAccountID      string
	ProviderAccountID    string
	ProviderAccountEmail string
	APIKeyCiphertext     []byte
	APIKeyNonce          []byte
	ProviderAPIKeyID     int64
	CredentialKeyVersion int
	Status               string
	BalanceUSD           float64
	ConcurrencyLimit     int
	AllowedModels        []string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ProviderAccount is the response from the AI provider account API.
type ProviderAccount struct {
	ProviderAccountID    string
	ProviderAccountEmail string
	UserID               int64
	Status               string
	BalanceUSD           float64
	ConcurrencyLimit     int
	AllowedModels        []string
}

// ProviderAPIKey is a one-time plaintext credential returned by the provider.
// Callers must immediately encrypt and persist it, then discard it from memory.
type ProviderAPIKey struct {
	ID        int64
	UserID    int64
	Key       string
	Name      string
	Status    string
	QuotaUSD  float64
	ExpiresAt *time.Time
}

// Credential is the decrypted provider credential used by server-side services.
type Credential struct {
	ProviderAccountID string
	APIKey            string
	KeyID             int64
}
