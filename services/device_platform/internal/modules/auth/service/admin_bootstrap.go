package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// BootstrapAdmin creates one initial administrator and enrolls TOTP.
//
// This operation is intended for a trusted local maintenance command. It
// returns the one-time otpauth URI so the operator can scan it into an
// authenticator; the secret is never persisted in plaintext or logged here.
func (s *Service) BootstrapAdmin(
	ctx context.Context,
	email string,
	password string,
	displayName string,
) (string, error) {
	email = normalizeEmail(email)
	displayName = strings.TrimSpace(displayName)
	if email == "" || !strings.Contains(email, "@") ||
		strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return "", domain.ErrInvalidEmail
	}
	if len(password) < 8 || len(password) > 128 ||
		!containsLetterAndNumber(password) {
		return "", domain.ErrWeakPassword
	}
	if displayName == "" || len([]rune(displayName)) > 40 {
		return "", domain.ErrInvalidDisplayName
	}

	existing, err := s.repository.GetParentAccountByEmail(ctx, email)
	if err == nil {
		if existing.Role != domain.RoleAdmin {
			return "", domain.ErrEmailExists
		}
		// A previous bootstrap may have created the account before TOTP
		// enrollment failed. Re-running only completes the missing enrollment.
		return s.EnrollAdminTOTP(ctx, existing.ID)
	}
	if !errors.Is(err, domain.ErrAccountNotFound) {
		return "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", fmt.Errorf("hash administrator password: %w", err)
	}
	now := s.timeSource.Now().UTC()
	accountID := uuid.NewString()
	account := &domain.ParentAccount{
		ID:           accountID,
		Email:        email,
		Phone:        "admin:" + accountID,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		Status:       accountStatusActive,
		Role:         domain.RoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repository.CreateParentAccount(ctx, account); err != nil {
		return "", err
	}
	uri, err := s.EnrollAdminTOTP(ctx, account.ID)
	if err != nil {
		return "", err
	}
	return uri, nil
}

// ResetAdminPassword replaces the password for one existing administrator.
//
// This operation is intended for a trusted local maintenance command. It
// deliberately refuses non-admin accounts and does not touch TOTP enrollment.
func (s *Service) ResetAdminPassword(
	ctx context.Context,
	email string,
	password string,
) error {
	email = normalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") ||
		strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return domain.ErrInvalidEmail
	}
	if len(password) < 8 || len(password) > 128 ||
		!containsLetterAndNumber(password) {
		return domain.ErrWeakPassword
	}

	account, err := s.repository.GetParentAccountByEmail(ctx, email)
	if err != nil {
		return err
	}
	if account.Role != domain.RoleAdmin {
		return domain.ErrInsufficientPrivilege
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("hash administrator password: %w", err)
	}
	return s.repository.UpdatePasswordHash(ctx, account.ID, string(passwordHash))
}
