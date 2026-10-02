// Package service contains parent authentication use cases.
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/repository"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/clock"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/security"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	accountStatusActive = "active"
	defaultConsent      = "2026-01"
)

// AIAccountProvisioner creates or repairs the parent's AI execution account.
// It must be idempotent for one parent account id.
type AIAccountProvisioner interface {
	EnsureForParent(
		ctx context.Context,
		parentAccountID string,
		parentEmail string,
	) (*domain.AIAccountSummary, error)
	GetForParent(
		ctx context.Context,
		parentAccountID string,
	) (*domain.AIAccountSummary, error)
}

// Service owns parent registration, login, token rotation, and account reads.
type Service struct {
	repository    repository.Repository
	tokenIssuer   security.TokenIssuer
	aiProvisioner AIAccountProvisioner
	timeSource    clock.Clock
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// Options contains authentication service dependencies and policies.
type Options struct {
	Repository    repository.Repository
	TokenIssuer   security.TokenIssuer
	AIProvisioner AIAccountProvisioner
	Clock         clock.Clock
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// New creates the parent authentication service.
func New(options Options) (*Service, error) {
	if options.Repository == nil {
		return nil, errors.New("authentication repository is required")
	}
	if options.TokenIssuer == nil {
		return nil, errors.New("token issuer is required")
	}
	if options.AccessTTL <= 0 || options.RefreshTTL <= 0 {
		return nil, errors.New("authentication token TTLs must be positive")
	}
	timeSource := options.Clock
	if timeSource == nil {
		timeSource = clock.SystemClock{}
	}
	return &Service{
		repository:    options.Repository,
		tokenIssuer:   options.TokenIssuer,
		aiProvisioner: options.AIProvisioner,
		timeSource:    timeSource,
		accessTTL:     options.AccessTTL,
		refreshTTL:    options.RefreshTTL,
	}, nil
}

// Register creates a guardian account and best-effort creates its AI account.
// A temporary AI gateway outage must not discard a valid parent registration;
// the AI projection remains unavailable until a later login or read repairs it.
func (s *Service) Register(
	ctx context.Context,
	input domain.RegisterInput,
) (*domain.ParentAccount, *domain.TokenPair, error) {
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.GuardianConsentVersion = strings.TrimSpace(input.GuardianConsentVersion)
	if input.GuardianConsentVersion == "" {
		input.GuardianConsentVersion = defaultConsent
	}
	if err := validateRegistration(input); err != nil {
		return nil, nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	now := s.timeSource.Now().UTC()
	account := &domain.ParentAccount{
		ID:                     uuid.NewString(),
		Email:                  input.Email,
		Phone:                  input.Phone,
		PasswordHash:           string(passwordHash),
		DisplayName:            input.DisplayName,
		Status:                 accountStatusActive,
		Role:                   domain.RoleParent,
		GuardianConsentVersion: input.GuardianConsentVersion,
		GuardianConsentedAt:    now,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := s.repository.CreateParentAccount(ctx, account); err != nil {
		return nil, nil, err
	}

	tokenPair, err := s.issueSession(ctx, account.ID)
	if err != nil {
		return nil, nil, err
	}
	if s.aiProvisioner != nil {
		// AI provisioning is deliberately after account and session creation:
		// the parent account remains recoverable when the AI gateway is down.
		_, _ = s.aiProvisioner.EnsureForParent(ctx, account.ID, account.Email)
	}
	return account, tokenPair, nil
}

// Login validates credentials and creates a new renewable session.
func (s *Service) Login(
	ctx context.Context,
	input domain.LoginInput,
) (*domain.ParentAccount, *domain.TokenPair, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, nil, domain.ErrInvalidCredentials
	}
	account, err := s.repository.GetParentAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, nil, domain.ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if account.Status != accountStatusActive {
		return nil, nil, domain.ErrAccountDisabled
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(account.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := s.repository.UpdateLastLogin(ctx, account.ID); err != nil {
		return nil, nil, err
	}
	tokenPair, err := s.issueSession(ctx, account.ID)
	if err != nil {
		return nil, nil, err
	}
	if s.aiProvisioner != nil {
		_, _ = s.aiProvisioner.EnsureForParent(ctx, account.ID, account.Email)
	}
	return account, tokenPair, nil
}

// Refresh rotates a refresh token and returns a fresh access token.
func (s *Service) Refresh(
	ctx context.Context,
	refreshToken string,
) (*domain.TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, domain.ErrSessionNotFound
	}
	session, err := s.repository.GetSessionByRefreshTokenHash(
		ctx,
		hashRefreshToken(refreshToken),
	)
	if err != nil {
		return nil, err
	}
	if !session.ExpiresAt.After(s.timeSource.Now().UTC()) {
		_ = s.repository.RevokeSession(ctx, session.ID)
		return nil, domain.ErrSessionExpired
	}
	account, err := s.repository.GetParentAccountByID(ctx, session.ParentAccountID)
	if err != nil {
		return nil, err
	}
	if account.Status != accountStatusActive {
		return nil, domain.ErrAccountDisabled
	}

	newRefreshToken, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	expiresAt := s.timeSource.Now().UTC().Add(s.refreshTTL)
	if err := s.repository.RotateSession(
		ctx,
		session.ID,
		hashRefreshToken(newRefreshToken),
		expiresAt,
	); err != nil {
		return nil, err
	}
	accessToken, err := s.issueAccessToken(account.ID)
	if err != nil {
		return nil, err
	}
	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.accessTTL,
	}, nil
}

// Logout revokes one session. Repeated logout is intentionally successful.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	session, err := s.repository.GetSessionByRefreshTokenHash(
		ctx,
		hashRefreshToken(refreshToken),
	)
	if errors.Is(err, domain.ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.repository.RevokeSession(ctx, session.ID)
}

// GetAccount returns the authenticated parent and parent-safe AI summary.
func (s *Service) GetAccount(
	ctx context.Context,
	accountID string,
) (*domain.ParentAccount, *domain.AIAccountSummary, error) {
	account, err := s.repository.GetParentAccountByID(ctx, accountID)
	if err != nil {
		return nil, nil, err
	}
	if account.Status != accountStatusActive {
		return nil, nil, domain.ErrAccountDisabled
	}
	if s.aiProvisioner == nil {
		return account, nil, nil
	}
	summary, err := s.aiProvisioner.EnsureForParent(ctx, account.ID, account.Email)
	if err != nil {
		// A provider outage is reported as an unavailable AI summary rather
		// than turning a valid parent login into an authentication failure.
		return account, nil, nil
	}
	return account, summary, nil
}

// VerifyAccessToken returns the subject for a valid bearer access token.
func (s *Service) VerifyAccessToken(accessToken string) (string, error) {
	claims, err := s.tokenIssuer.Verify(accessToken)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func (s *Service) issueSession(
	ctx context.Context,
	parentAccountID string,
) (*domain.TokenPair, error) {
	accessToken, err := s.issueAccessToken(parentAccountID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	now := s.timeSource.Now().UTC()
	session := &domain.Session{
		ID:               uuid.NewString(),
		ParentAccountID:  parentAccountID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        now.Add(s.refreshTTL),
		CreatedAt:        now,
		LastUsedAt:       now,
	}
	if err := s.repository.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.accessTTL,
	}, nil
}

func (s *Service) issueAccessToken(parentAccountID string) (string, error) {
	now := s.timeSource.Now().UTC()
	return s.tokenIssuer.Issue(security.TokenClaims{
		Subject:   parentAccountID,
		TokenID:   uuid.NewString(),
		IssuedAt:  now,
		ExpiresAt: now.Add(s.accessTTL),
	})
}

func validateRegistration(input domain.RegisterInput) error {
	if input.Email == "" || !strings.Contains(input.Email, "@") ||
		strings.HasPrefix(input.Email, "@") || strings.HasSuffix(input.Email, "@") {
		return domain.ErrInvalidEmail
	}
	if len(input.Password) < 8 || len(input.Password) > 128 ||
		!containsLetterAndNumber(input.Password) {
		return domain.ErrWeakPassword
	}
	if input.DisplayName == "" || len([]rune(input.DisplayName)) > 40 {
		return domain.ErrInvalidDisplayName
	}
	if input.GuardianConsentVersion == "" {
		return domain.ErrGuardianConsent
	}
	return nil
}

func containsLetterAndNumber(value string) bool {
	hasLetter := false
	hasNumber := false
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z':
			hasLetter = true
		case character >= 'A' && character <= 'Z':
			hasLetter = true
		case character >= '0' && character <= '9':
			hasNumber = true
		}
	}
	return hasLetter && hasNumber
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func randomToken() (string, error) {
	bytes := make([]byte, 48)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashRefreshToken(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
