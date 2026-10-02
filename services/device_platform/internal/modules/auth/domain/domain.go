// Package domain contains parent authentication domain types.
package domain

import (
	"errors"
	"time"
)

var (
	ErrEmailExists           = errors.New("email already exists")
	ErrAccountNotFound       = errors.New("parent account not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountDisabled       = errors.New("parent account is disabled")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionExpired        = errors.New("session expired")
	ErrGuardianConsent       = errors.New("guardian consent is required")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrWeakPassword          = errors.New("password does not meet requirements")
	ErrInvalidDisplayName    = errors.New("display name is required")
	ErrAIAccountUnavailable  = errors.New("AI account is unavailable")
	ErrInsufficientPrivilege = errors.New("insufficient privilege")
)

const (
	RoleParent = "parent"
	RoleAdmin  = "admin"
)

// ParentAccount is the guardian identity that owns family, device, and AI data.
type ParentAccount struct {
	ID                     string
	Email                  string
	Phone                  string
	PasswordHash           string
	DisplayName            string
	Status                 string
	GuardianConsentVersion string
	GuardianConsentedAt    time.Time
	Role                   string
	LastLoginAt            *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// Session is one renewable parent login session.
type Session struct {
	ID               string
	ParentAccountID  string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
	LastUsedAt       time.Time
}

// TokenPair is the transport result for login and refresh operations.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
}

// RegisterInput contains the minimum data allowed during parent registration.
type RegisterInput struct {
	Email                  string
	Password               string
	DisplayName            string
	Phone                  string
	GuardianConsentVersion string
}

// LoginInput contains credentials for one parent login attempt.
type LoginInput struct {
	Email    string
	Password string
}

// AuthenticatedParent is returned to authorized platform modules.
type AuthenticatedParent struct {
	AccountID   string
	Email       string
	DisplayName string
	Role        string
}

// AIAccountSummary is the parent-safe projection of the linked AI account.
// It never contains a provider API key or an internal service identifier.
type AIAccountSummary struct {
	Status           string
	BalanceUSD       float64
	ConcurrencyLimit int
	AllowedModels    []string
	ProviderReady    bool
}
