// Package repository persists platform-side AI account projections.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository stores parent AI account state and encrypted credentials.
type Repository interface {
	GetByParentAccountID(
		ctx context.Context,
		parentAccountID string,
	) (*domain.Account, error)
	GetByProviderAccountID(
		ctx context.Context,
		providerAccountID string,
	) (*domain.Account, error)
	List(ctx context.Context) ([]domain.Account, error)
	Create(ctx context.Context, account *domain.Account) error
	UpdateFromProvider(ctx context.Context, account *domain.Account) error
	UpdateStatus(
		ctx context.Context,
		parentAccountID string,
		status string,
	) error
}

// PostgresRepository is the PostgreSQL-backed AI account repository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates an AI account repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetByParentAccountID loads one account by the platform parent id.
func (r *PostgresRepository) GetByParentAccountID(
	ctx context.Context,
	parentAccountID string,
) (*domain.Account, error) {
	return r.getOne(ctx, `
		SELECT
			id,
			parent_account_id,
			provider_account_id,
			provider_account_email,
			credential_ciphertext,
			credential_nonce,
			provider_api_key_id,
			credential_key_version,
			status,
			balance_usd,
			concurrency_limit,
			allowed_models,
			created_at,
			updated_at
		FROM ai_accounts
		WHERE parent_account_id = $1
	`, parentAccountID)
}

// GetByProviderAccountID loads one account by the provider namespace id.
func (r *PostgresRepository) GetByProviderAccountID(
	ctx context.Context,
	providerAccountID string,
) (*domain.Account, error) {
	return r.getOne(ctx, `
		SELECT
			id,
			parent_account_id,
			provider_account_id,
			provider_account_email,
			credential_ciphertext,
			credential_nonce,
			provider_api_key_id,
			credential_key_version,
			status,
			balance_usd,
			concurrency_limit,
			allowed_models,
			created_at,
			updated_at
		FROM ai_accounts
		WHERE provider_account_id = $1
	`, providerAccountID)
}

// List returns all platform AI account projections for the operations API.
func (r *PostgresRepository) List(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			parent_account_id,
			provider_account_id,
			provider_account_email,
			credential_ciphertext,
			credential_nonce,
			provider_api_key_id,
			credential_key_version,
			status,
			balance_usd,
			concurrency_limit,
			allowed_models,
			created_at,
			updated_at
		FROM ai_accounts
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list AI accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0)
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI accounts: %w", err)
	}
	return accounts, nil
}

// Create inserts the platform account projection.
func (r *PostgresRepository) Create(ctx context.Context, account *domain.Account) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_accounts (
			id,
			parent_account_id,
			provider_account_id,
			provider_account_email,
			credential_ciphertext,
			credential_nonce,
			provider_api_key_id,
			credential_key_version,
			status,
			balance_usd,
			concurrency_limit,
			allowed_models,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		account.ID,
		account.ParentAccountID,
		account.ProviderAccountID,
		account.ProviderAccountEmail,
		account.APIKeyCiphertext,
		account.APIKeyNonce,
		account.ProviderAPIKeyID,
		account.CredentialKeyVersion,
		account.Status,
		account.BalanceUSD,
		account.ConcurrencyLimit,
		account.AllowedModels,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert AI account: %w", err)
	}
	return nil
}

// UpdateFromProvider updates the non-secret provider projection without
// touching credential material.
func (r *PostgresRepository) UpdateFromProvider(
	ctx context.Context,
	account *domain.Account,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ai_accounts
		SET status = $2,
		    balance_usd = $3,
		    concurrency_limit = $4,
		    allowed_models = $5,
		    credential_ciphertext = $6,
		    credential_nonce = $7,
		    provider_api_key_id = $8,
		    credential_key_version = $9,
		    updated_at = $10
		WHERE id = $1
	`,
		account.ID,
		account.Status,
		account.BalanceUSD,
		account.ConcurrencyLimit,
		account.AllowedModels,
		account.APIKeyCiphertext,
		account.APIKeyNonce,
		account.ProviderAPIKeyID,
		account.CredentialKeyVersion,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update AI account: %w", err)
	}
	return nil
}

// UpdateStatus changes only the platform-side lifecycle state.
func (r *PostgresRepository) UpdateStatus(
	ctx context.Context,
	parentAccountID string,
	status string,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ai_accounts
		SET status = $2,
		    updated_at = NOW()
		WHERE parent_account_id = $1
	`, parentAccountID, status)
	if err != nil {
		return fmt.Errorf("update AI account status: %w", err)
	}
	return nil
}

func (r *PostgresRepository) getOne(
	ctx context.Context,
	query string,
	argument any,
) (*domain.Account, error) {
	row := r.pool.QueryRow(ctx, query, argument)
	return scanAccount(row)
}

type accountScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row accountScanner) (*domain.Account, error) {
	var account domain.Account
	err := row.Scan(
		&account.ID,
		&account.ParentAccountID,
		&account.ProviderAccountID,
		&account.ProviderAccountEmail,
		&account.APIKeyCiphertext,
		&account.APIKeyNonce,
		&account.ProviderAPIKeyID,
		&account.CredentialKeyVersion,
		&account.Status,
		&account.BalanceUSD,
		&account.ConcurrencyLimit,
		&account.AllowedModels,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get AI account: %w", err)
	}
	return &account, nil
}
