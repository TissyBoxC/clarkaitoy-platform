// Package repository persists device binding tokens and durable bindings.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines device binding persistence operations.
type Repository interface {
	CreateToken(ctx context.Context, token *domain.BindingToken) error
	GetTokenByHash(ctx context.Context, tokenHash string) (*domain.BindingToken, error)
	ConsumeToken(ctx context.Context, tokenID string) error
	UpsertBinding(ctx context.Context, binding *domain.Binding) error
	ListByParentAccountID(
		ctx context.Context,
		parentAccountID string,
	) ([]domain.Binding, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*domain.Binding, error)
	Delete(ctx context.Context, parentAccountID string, deviceID string) error
}

// PostgresRepository is the PostgreSQL-backed device binding repository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a device binding repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateToken inserts a hashed, single-use provisioning token.
func (r *PostgresRepository) CreateToken(
	ctx context.Context,
	token *domain.BindingToken,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO device_binding_tokens (
			id,
			device_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		token.ID,
		token.DeviceID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert device binding token: %w", err)
	}
	return nil
}

// GetTokenByHash loads one unconsumed token by its hash.
func (r *PostgresRepository) GetTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*domain.BindingToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			device_id,
			token_hash,
			expires_at,
			consumed_at,
			created_at
		FROM device_binding_tokens
		WHERE token_hash = $1
	`, tokenHash)

	var token domain.BindingToken
	err := row.Scan(
		&token.ID,
		&token.DeviceID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device binding token: %w", err)
	}
	return &token, nil
}

// ConsumeToken atomically marks a token as used.
func (r *PostgresRepository) ConsumeToken(
	ctx context.Context,
	tokenID string,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE device_binding_tokens
		SET consumed_at = NOW()
		WHERE id = $1
		  AND consumed_at IS NULL
		  AND expires_at > NOW()
	`, tokenID)
	if err != nil {
		return fmt.Errorf("consume device binding token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTokenConsumed
	}
	return nil
}

// UpsertBinding binds or transfers one device to a parent account.
func (r *PostgresRepository) UpsertBinding(
	ctx context.Context,
	binding *domain.Binding,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO device_bindings (
			id,
			parent_account_id,
			device_id,
			device_name,
			hardware_model,
			firmware_version,
			capability_set,
			bound_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (device_id) DO UPDATE
		SET parent_account_id = EXCLUDED.parent_account_id,
		    device_name = EXCLUDED.device_name,
		    hardware_model = EXCLUDED.hardware_model,
		    firmware_version = EXCLUDED.firmware_version,
		    capability_set = EXCLUDED.capability_set,
		    updated_at = EXCLUDED.updated_at
	`,
		binding.ID,
		binding.ParentAccountID,
		binding.DeviceID,
		binding.DeviceName,
		binding.HardwareModel,
		binding.FirmwareVersion,
		binding.Capabilities,
		binding.BoundAt,
		binding.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert device binding: %w", err)
	}
	return nil
}

// ListByParentAccountID returns all devices owned by one parent.
func (r *PostgresRepository) ListByParentAccountID(
	ctx context.Context,
	parentAccountID string,
) ([]domain.Binding, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			parent_account_id,
			device_id,
			device_name,
			hardware_model,
			firmware_version,
			capability_set,
			bound_at,
			updated_at
		FROM device_bindings
		WHERE parent_account_id = $1
		ORDER BY updated_at DESC
	`, parentAccountID)
	if err != nil {
		return nil, fmt.Errorf("list device bindings: %w", err)
	}
	defer rows.Close()

	bindings := make([]domain.Binding, 0)
	for rows.Next() {
		binding, err := scanBinding(rows)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, *binding)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device bindings: %w", err)
	}
	return bindings, nil
}

// GetByDeviceID loads the current owner of one device.
func (r *PostgresRepository) GetByDeviceID(
	ctx context.Context,
	deviceID string,
) (*domain.Binding, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			parent_account_id,
			device_id,
			device_name,
			hardware_model,
			firmware_version,
			capability_set,
			bound_at,
			updated_at
		FROM device_bindings
		WHERE device_id = $1
	`, deviceID)
	return scanBinding(row)
}

// Delete removes one binding only when it belongs to the requesting parent.
func (r *PostgresRepository) Delete(
	ctx context.Context,
	parentAccountID string,
	deviceID string,
) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM device_bindings
		WHERE parent_account_id = $1
		  AND device_id = $2
	`, parentAccountID, deviceID)
	if err != nil {
		return fmt.Errorf("delete device binding: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDeviceNotFound
	}
	return nil
}

type bindingScanner interface {
	Scan(dest ...any) error
}

func scanBinding(row bindingScanner) (*domain.Binding, error) {
	var binding domain.Binding
	err := row.Scan(
		&binding.ID,
		&binding.ParentAccountID,
		&binding.DeviceID,
		&binding.DeviceName,
		&binding.HardwareModel,
		&binding.FirmwareVersion,
		&binding.Capabilities,
		&binding.BoundAt,
		&binding.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrDeviceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan device binding: %w", err)
	}
	return &binding, nil
}
