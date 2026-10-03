package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestResetParentPasswordUpdatesHash(t *testing.T) {
	repository := &memoryRepository{
		account: &domain.ParentAccount{
			ID:           "parent-001",
			Role:         domain.RoleParent,
			PasswordHash: mustPasswordHash(t, "old-password"),
			Status:       accountStatusActive,
		},
	}
	service := newAdminTestService(t, repository)

	if err := service.ResetParentPassword(
		context.Background(),
		"parent-001",
		"new-password1",
	); err != nil {
		t.Fatalf("reset parent password: %v", err)
	}
	if repository.passwordHashUpdates != 1 {
		t.Fatalf("expected one password update, got %d", repository.passwordHashUpdates)
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(repository.account.PasswordHash),
		[]byte("new-password1"),
	); err != nil {
		t.Fatalf("stored password hash does not match the new password: %v", err)
	}
}

func TestResetParentPasswordRejectsWeakPasswordBeforeRepositoryAccess(t *testing.T) {
	repository := &memoryRepository{
		account: &domain.ParentAccount{
			ID:     "parent-001",
			Role:   domain.RoleParent,
			Status: accountStatusActive,
		},
	}
	service := newAdminTestService(t, repository)

	err := service.ResetParentPassword(context.Background(), "parent-001", "weak")
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected weak password, got %v", err)
	}
	if repository.passwordHashUpdates != 0 {
		t.Fatal("weak password must not reach the repository")
	}
}

func TestResetParentPasswordRejectsNonParent(t *testing.T) {
	repository := &memoryRepository{
		account: &domain.ParentAccount{
			ID:     "admin-001",
			Role:   domain.RoleAdmin,
			Status: accountStatusActive,
		},
	}
	service := newAdminTestService(t, repository)

	err := service.ResetParentPassword(
		context.Background(),
		"admin-001",
		"new-password1",
	)
	if !errors.Is(err, domain.ErrInsufficientPrivilege) {
		t.Fatalf("expected insufficient privilege, got %v", err)
	}
	if repository.passwordHashUpdates != 0 {
		t.Fatal("administrator password must not be changed by the parent reset path")
	}
}

func TestResetParentPasswordPropagatesMissingAccount(t *testing.T) {
	service := newAdminTestService(t, &memoryRepository{})

	err := service.ResetParentPassword(
		context.Background(),
		"missing-parent",
		"new-password1",
	)
	if !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("expected account not found, got %v", err)
	}
}

func TestUpdateParentProfilePersistsEditableFields(t *testing.T) {
	repository := &memoryRepository{
		account: &domain.ParentAccount{
			ID:           "parent-001",
			DisplayName:  "家长",
			Role:         domain.RoleParent,
			Status:       accountStatusActive,
			PasswordHash: mustPasswordHash(t, "password1"),
		},
	}
	service := newAdminTestService(t, repository)

	updated, err := service.UpdateParentProfile(
		context.Background(),
		"parent-001",
		ProfileUpdate{
			DisplayName:        "林家长",
			GuardianFamilyName: "林",
			ChildNickname:      "小芽",
			ChildBirthday:      "2020-05-20",
		},
	)
	if err != nil {
		t.Fatalf("update parent profile: %v", err)
	}
	if updated.DisplayName != "林家长" ||
		updated.GuardianFamilyName != "林" ||
		updated.ChildNickname != "小芽" ||
		updated.ChildBirthday != "2020-05-20" {
		t.Fatalf("unexpected profile after update: %+v", updated)
	}
	if repository.account.DisplayName != "林家长" {
		t.Fatalf("expected profile to be persisted, got %q", repository.account.DisplayName)
	}
}

func TestUpdateParentProfileRejectsNonParent(t *testing.T) {
	repository := &memoryRepository{
		account: &domain.ParentAccount{
			ID:     "admin-001",
			Role:   domain.RoleAdmin,
			Status: accountStatusActive,
		},
	}
	service := newAdminTestService(t, repository)

	_, err := service.UpdateParentProfile(
		context.Background(),
		"admin-001",
		ProfileUpdate{DisplayName: "管理员"},
	)
	if !errors.Is(err, domain.ErrInsufficientPrivilege) {
		t.Fatalf("expected insufficient privilege, got %v", err)
	}
}

func TestUpdateParentProfilePropagatesMissingAccount(t *testing.T) {
	service := newAdminTestService(t, &memoryRepository{})

	_, err := service.UpdateParentProfile(
		context.Background(),
		"missing-parent",
		ProfileUpdate{DisplayName: "家长"},
	)
	if !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("expected account not found, got %v", err)
	}
}

func newAdminTestService(t *testing.T, repository *memoryRepository) *Service {
	t.Helper()
	service, err := New(Options{
		Repository:  repository,
		TokenIssuer: mustTokenIssuer(t),
		AccessTTL:   time.Minute,
		RefreshTTL:  time.Hour,
	})
	if err != nil {
		t.Fatalf("create authentication service: %v", err)
	}
	return service
}
