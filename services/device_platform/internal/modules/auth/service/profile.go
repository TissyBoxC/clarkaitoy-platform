package service

import (
	"context"
	"strings"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
)

// ProfileUpdate contains only fields a guardian may edit after registration.
// Phone and role are immutable because they control identity and authorization.
type ProfileUpdate struct {
	DisplayName        string
	GuardianFamilyName string
	ChildNickname      string
	ChildBirthday      string
}

// UpdateProfile updates guardian-editable profile fields for one account.
func (s *Service) UpdateProfile(
	ctx context.Context,
	accountID string,
	update ProfileUpdate,
) (*domain.ParentAccount, error) {
	account, err := s.GetIdentity(ctx, accountID)
	if err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(update.DisplayName)
	if displayName == "" {
		displayName = guardianDisplayName(domain.RegisterInput{
			GuardianFamilyName: update.GuardianFamilyName,
			ChildNickname:      update.ChildNickname,
		})
	}
	if len([]rune(displayName)) > 40 {
		return nil, domain.ErrInvalidDisplayName
	}
	if len([]rune(strings.TrimSpace(update.GuardianFamilyName))) > 40 {
		return nil, domain.ErrInvalidGuardianName
	}
	if len([]rune(strings.TrimSpace(update.ChildNickname))) > 40 {
		return nil, domain.ErrInvalidChildNickname
	}
	childBirthday := strings.TrimSpace(update.ChildBirthday)
	if childBirthday != "" {
		birthday, err := time.Parse("2006-01-02", childBirthday)
		if err != nil || birthday.After(s.timeSource.Now().UTC()) {
			return nil, domain.ErrInvalidChildBirthday
		}
	}
	account.DisplayName = displayName
	account.GuardianFamilyName = strings.TrimSpace(update.GuardianFamilyName)
	account.ChildNickname = strings.TrimSpace(update.ChildNickname)
	account.ChildBirthday = childBirthday
	account.UpdatedAt = s.timeSource.Now().UTC()
	if err := s.repository.UpdateProfile(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

// UpdateParentProfile updates guardian-editable fields for one parent account.
//
// The administrator-facing entry point refuses non-parent roles before
// delegating to the shared profile validation and persistence path.
func (s *Service) UpdateParentProfile(
	ctx context.Context,
	accountID string,
	update ProfileUpdate,
) (*domain.ParentAccount, error) {
	account, err := s.GetIdentity(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Role != domain.RoleParent {
		return nil, domain.ErrInsufficientPrivilege
	}
	return s.UpdateProfile(ctx, accountID, update)
}
