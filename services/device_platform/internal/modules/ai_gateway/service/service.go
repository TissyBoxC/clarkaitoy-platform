// Package service owns parent AI account provisioning and credential lifecycle.
package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/repository"
	authdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/clock"
	"github.com/google/uuid"
)

const (
	statusActive    = "active"
	statusSuspended = "suspended"
)

// Provider calls the private sub2api account and API key API.
type Provider interface {
	GetAccount(
		ctx context.Context,
		providerAccountID string,
	) (*domain.ProviderAccount, error)
	CreateAccount(
		ctx context.Context,
		request domain.ProviderAccount,
		password string,
	) (*domain.ProviderAccount, error)
	UpdateAccount(
		ctx context.Context,
		providerAccountID string,
		request domain.ProviderAccount,
		reason string,
	) (*domain.ProviderAccount, error)
	CreateAPIKey(
		ctx context.Context,
		providerAccountID string,
		request domain.ProviderAPIKey,
	) (*domain.ProviderAPIKey, error)
	RotateAPIKey(
		ctx context.Context,
		providerAccountID string,
		apiKeyID int64,
		request domain.ProviderAPIKey,
	) (*domain.ProviderAPIKey, error)
	DeleteAPIKey(
		ctx context.Context,
		providerAccountID string,
		apiKeyID int64,
	) error
}

// CredentialCipher encrypts provider API keys before they reach PostgreSQL.
type CredentialCipher interface {
	Encrypt(plaintext []byte) (ciphertext []byte, nonce []byte, err error)
	Decrypt(ciphertext []byte, nonce []byte) ([]byte, error)
}

// Service implements auth.AIAccountProvisioner and administrative AI flows.
type Service struct {
	repository         repository.Repository
	provider           Provider
	cipher             CredentialCipher
	timeSource         clock.Clock
	defaultBalance     float64
	defaultModels      []string
	defaultConcurrency int
}

// Options contains AI account service dependencies and defaults.
type Options struct {
	Repository         repository.Repository
	Provider           Provider
	Cipher             CredentialCipher
	Clock              clock.Clock
	DefaultBalanceUSD  float64
	DefaultModels      []string
	DefaultConcurrency int
}

// New creates the AI account service.
func New(options Options) (*Service, error) {
	if options.Repository == nil {
		return nil, errors.New("AI account repository is required")
	}
	if options.Provider == nil {
		return nil, errors.New("AI account provider is required")
	}
	if options.Cipher == nil {
		return nil, errors.New("AI credential cipher is required")
	}
	if options.DefaultConcurrency < 1 {
		return nil, errors.New("AI default concurrency must be positive")
	}
	timeSource := options.Clock
	if timeSource == nil {
		timeSource = clock.SystemClock{}
	}
	return &Service{
		repository:         options.Repository,
		provider:           options.Provider,
		cipher:             options.Cipher,
		timeSource:         timeSource,
		defaultBalance:     options.DefaultBalanceUSD,
		defaultModels:      append([]string(nil), options.DefaultModels...),
		defaultConcurrency: options.DefaultConcurrency,
	}, nil
}

// EnsureForParent creates or repairs the account projection and API key.
//
// The operation is idempotent. If the provider created a key but the platform
// failed before persisting it, a later retry rotates the existing provider key
// rather than accumulating valid credentials.
func (s *Service) EnsureForParent(
	ctx context.Context,
	parentAccountID string,
	parentEmail string,
) (*authdomain.AIAccountSummary, error) {
	existing, err := s.repository.GetByParentAccountID(ctx, parentAccountID)
	if err != nil && !errors.Is(err, domain.ErrAccountNotFound) {
		return nil, err
	}
	if existing != nil && len(existing.APIKeyCiphertext) > 0 &&
		len(existing.APIKeyNonce) > 0 {
		return s.summaryFromAccount(existing), nil
	}

	providerAccountID := providerAccountIDForParent(parentAccountID)
	providerAccount, err := s.provider.GetAccount(ctx, providerAccountID)
	if errors.Is(err, domain.ErrAccountNotFound) {
		providerAccount, err = s.provider.CreateAccount(
			ctx,
			domain.ProviderAccount{
				ProviderAccountID:    providerAccountID,
				ProviderAccountEmail: controlledEmail(parentEmail, parentAccountID),
				Status:               statusActive,
				BalanceUSD:           s.defaultBalance,
				ConcurrencyLimit:     s.defaultConcurrency,
				AllowedModels:        append([]string(nil), s.defaultModels...),
			},
			randomProviderPassword(),
		)
	}
	if err != nil {
		return nil, err
	}
	if providerAccount == nil {
		return nil, domain.ErrProviderUnavailable
	}

	existingKeyID := int64(0)
	if existing != nil {
		existingKeyID = existing.ProviderAPIKeyID
	}
	providerKey, err := s.ensureKey(ctx, providerAccountID, existingKeyID)
	if err != nil {
		return nil, err
	}
	encryptedKey, nonce, err := s.cipher.Encrypt([]byte(providerKey.Key))
	if err != nil {
		return nil, fmt.Errorf("encrypt AI credential: %w", err)
	}

	now := s.timeSource.Now().UTC()
	if existing == nil {
		existing = &domain.Account{
			ID:                   uuid.NewString(),
			ParentAccountID:      parentAccountID,
			ProviderAccountID:    providerAccountID,
			ProviderAccountEmail: controlledEmail(parentEmail, parentAccountID),
			Status:               providerAccount.Status,
			BalanceUSD:           providerAccount.BalanceUSD,
			ConcurrencyLimit:     providerAccount.ConcurrencyLimit,
			AllowedModels:        providerAccount.AllowedModels,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := s.repository.Create(ctx, existing); err != nil {
			// Do not leave a usable provider credential if the platform could
			// not persist its encrypted copy.
			_ = s.provider.DeleteAPIKey(ctx, providerAccountID, providerKey.ID)
			return nil, err
		}
	}
	existing.APIKeyCiphertext = encryptedKey
	existing.APIKeyNonce = nonce
	existing.ProviderAPIKeyID = providerKey.ID
	existing.CredentialKeyVersion = 1
	existing.Status = providerAccount.Status
	existing.BalanceUSD = providerAccount.BalanceUSD
	existing.ConcurrencyLimit = providerAccount.ConcurrencyLimit
	existing.AllowedModels = providerAccount.AllowedModels
	existing.UpdatedAt = now
	if err := s.repository.UpdateFromProvider(ctx, existing); err != nil {
		return nil, err
	}
	return s.summaryFromAccount(existing), nil
}

// GetForParent returns the parent-safe AI account summary.
func (s *Service) GetForParent(
	ctx context.Context,
	parentAccountID string,
) (*authdomain.AIAccountSummary, error) {
	account, err := s.repository.GetByParentAccountID(ctx, parentAccountID)
	if err != nil {
		return nil, err
	}
	return s.summaryFromAccount(account), nil
}

// ListForAdmin returns all platform AI account projections for operations.
func (s *Service) ListForAdmin(ctx context.Context) ([]domain.Account, error) {
	return s.repository.List(ctx)
}

// UpdateForAdmin applies an explicit admin change to provider and platform.
func (s *Service) UpdateForAdmin(
	ctx context.Context,
	providerAccountID string,
	status string,
	balanceUSD float64,
	concurrencyLimit int,
	allowedModels []string,
	reason string,
) (*domain.Account, error) {
	account, err := s.repository.GetByProviderAccountID(ctx, providerAccountID)
	if err != nil {
		return nil, err
	}
	providerAccount, err := s.provider.UpdateAccount(
		ctx,
		providerAccountID,
		domain.ProviderAccount{
			ProviderAccountID: providerAccountID,
			Status:            status,
			BalanceUSD:        balanceUSD,
			ConcurrencyLimit:  concurrencyLimit,
			AllowedModels:     append([]string(nil), allowedModels...),
		},
		reason,
	)
	if err != nil {
		return nil, err
	}
	account.Status = providerAccount.Status
	account.BalanceUSD = providerAccount.BalanceUSD
	account.ConcurrencyLimit = providerAccount.ConcurrencyLimit
	account.AllowedModels = providerAccount.AllowedModels
	account.UpdatedAt = s.timeSource.Now().UTC()
	if err := s.repository.UpdateFromProvider(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

// CredentialForDevice returns an internal credential for server-side relay use.
// It must never be exposed through parent, admin, firmware, or public APIs.
func (s *Service) CredentialForDevice(
	ctx context.Context,
	parentAccountID string,
) (*domain.Credential, error) {
	account, err := s.repository.GetByParentAccountID(ctx, parentAccountID)
	if err != nil {
		return nil, err
	}
	if account.Status != statusActive {
		return nil, domain.ErrCredentialInvalid
	}
	plaintext, err := s.cipher.Decrypt(
		account.APIKeyCiphertext,
		account.APIKeyNonce,
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt AI credential: %w", err)
	}
	return &domain.Credential{
		ProviderAccountID: account.ProviderAccountID,
		APIKey:            string(plaintext),
	}, nil
}

func (s *Service) ensureKey(
	ctx context.Context,
	providerAccountID string,
	existingKeyID int64,
) (*domain.ProviderAPIKey, error) {
	request := domain.ProviderAPIKey{
		Name:     "sprout-platform",
		QuotaUSD: 0,
	}
	if existingKeyID <= 0 {
		return s.provider.CreateAPIKey(ctx, providerAccountID, request)
	}
	return s.provider.RotateAPIKey(
		ctx,
		providerAccountID,
		existingKeyID,
		request,
	)
}

func (s *Service) summaryFromAccount(
	account *domain.Account,
) *authdomain.AIAccountSummary {
	return &authdomain.AIAccountSummary{
		Status:           account.Status,
		BalanceUSD:       account.BalanceUSD,
		ConcurrencyLimit: account.ConcurrencyLimit,
		AllowedModels:    append([]string(nil), account.AllowedModels...),
		ProviderReady:    len(account.APIKeyCiphertext) > 0 && len(account.APIKeyNonce) > 0,
	}
}

// AESGCMCipher encrypts credentials with a versioned, authenticated key.
type AESGCMCipher struct {
	aead cipher.AEAD
}

// NewAESGCMCipher derives a 256-bit key from configuration material.
func NewAESGCMCipher(secret string) (*AESGCMCipher, error) {
	if len(strings.TrimSpace(secret)) < 32 {
		return nil, errors.New("AI credential key must contain at least 32 characters")
	}
	derivedKey := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, fmt.Errorf("create credential cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create credential AEAD: %w", err)
	}
	return &AESGCMCipher{aead: aead}, nil
}

// Encrypt returns ciphertext and a unique nonce.
func (c *AESGCMCipher) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	if c == nil || c.aead == nil {
		return nil, nil, errors.New("credential cipher is not configured")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate credential nonce: %w", err)
	}
	return c.aead.Seal(nil, nonce, plaintext, nil), nonce, nil
}

// Decrypt authenticates and decrypts stored credential material.
func (c *AESGCMCipher) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	if c == nil || c.aead == nil {
		return nil, errors.New("credential cipher is not configured")
	}
	if len(nonce) != c.aead.NonceSize() {
		return nil, errors.New("credential nonce has an invalid length")
	}
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("credential authentication failed")
	}
	return plaintext, nil
}

func providerAccountIDForParent(parentAccountID string) string {
	compact := strings.ReplaceAll(strings.ToLower(parentAccountID), "-", "")
	return "parent_" + compact
}

func controlledEmail(parentEmail string, parentAccountID string) string {
	compact := strings.ReplaceAll(strings.ToLower(parentAccountID), "-", "")
	if len(compact) > 20 {
		compact = compact[:20]
	}
	return "parent." + compact + "@ai.sprout.local"
}

func randomProviderPassword() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// The caller will surface the provider failure; this path cannot safely
		// continue because a weak provider password would be created.
		panic("AI provider password source unavailable")
	}
	return "sp-" + hex.EncodeToString(bytes)
}
