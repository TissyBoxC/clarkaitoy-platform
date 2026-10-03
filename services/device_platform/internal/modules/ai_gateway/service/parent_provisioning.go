package service

import (
	"context"

	authdomain "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
)

// CreateParentWithAIAccount creates a guardian account and immediately
// provisions its dependent AI account.
//
// The parent record is the source of truth. If AI provisioning fails after the
// parent row is committed, the returned account is still valid and the
// administrator can retry provisioning without creating a duplicate parent.
// This is preferable to an open transaction that holds a database lock while
// calling an external provider.
func (s *Service) CreateParentWithAIAccount(
	ctx context.Context,
	authService parentCreator,
	input authdomain.RegisterInput,
) (*authdomain.ParentAccount, *authdomain.AIAccountSummary, error) {
	account, summary, err := authService.CreateParent(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if summary != nil {
		return account, summary, nil
	}
	summary, err = s.EnsureForParent(ctx, account.ID, account.Email)
	if err != nil {
		return account, nil, err
	}
	return account, summary, nil
}

// RepairParentAIAccount retries AI provisioning for one existing guardian.
// The parent identifier is resolved by the caller so this operation cannot
// create a second platform account or a second provider credential.
func (s *Service) RepairParentAIAccount(
	ctx context.Context,
	parentAccountID string,
	parentEmail string,
) (*authdomain.AIAccountSummary, error) {
	return s.EnsureForParent(ctx, parentAccountID, parentEmail)
}

type parentCreator interface {
	CreateParent(
		ctx context.Context,
		input authdomain.RegisterInput,
	) (*authdomain.ParentAccount, *authdomain.AIAccountSummary, error)
}
