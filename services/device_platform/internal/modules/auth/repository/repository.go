// Package repository persists parent accounts and refresh sessions.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for parent authentication.
type Repository interface {
	CreateParentAccount(ctx context.Context, account *domain.ParentAccount) error
	GetParentAccountByEmail(ctx context.Context, email string) (*domain.ParentAccount, error)
	GetParentAccountByID(ctx context.Context, accountID string) (*domain.ParentAccount, error)
	UpdateLastLogin(ctx context.Context, accountID string) error
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByRefreshTokenHash(
		ctx context.Context,
		refreshTokenHash string,
	) (*domain.Session, error)
	RotateSession(
		ctx context.Context,
		sessionID string,
		refreshTokenHash string,
		expiresAt time.Time,
	) error
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllSessions(ctx context.Context, parentAccountID string) error
}

// PostgresRepository is the PostgreSQL-backed authentication repository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates an authentication repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateParentAccount inserts one guardian account.
func (r *PostgresRepository) CreateParentAccount(
	ctx context.Context,
	account *domain.ParentAccount,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO parent_accounts (
			id,
			email,
			phone,
			password_hash,
			display_name,
			status,
			role,
			guardian_consent_version,
			guardian_consented_at,
			last_login_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		account.ID,
		account.Email,
		account.Phone,
		account.PasswordHash,
		account.DisplayName,
		account.Status,
		account.Role,
		account.GuardianConsentVersion,
		account.GuardianConsentedAt,
		account.LastLoginAt,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return domain.ErrEmailExists
		}
		return fmt.Errorf("insert parent account: %w", err)
	}
	return nil
}

// GetParentAccountByEmail loads one account by normalized email.
func (r *PostgresRepository) GetParentAccountByEmail(
	ctx context.Context,
	email string,
) (*domain.ParentAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			email,
			COALESCE(phone, ''),
			password_hash,
			display_name,
			status,
			role,
			guardian_consent_version,
			guardian_consented_at,
			last_login_at,
			created_at,
			updated_at
		FROM parent_accounts
		WHERE email = $1
	`, email)
	return scanParentAccount(row)
}

// GetParentAccountByID loads one account by stable identifier.
func (r *PostgresRepository) GetParentAccountByID(
	ctx context.Context,
	accountID string,
) (*domain.ParentAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			email,
			COALESCE(phone, ''),
			password_hash,
			display_name,
			status,
			role,
			guardian_consent_version,
			guardian_consented_at,
			last_login_at,
			created_at,
			updated_at
		FROM parent_accounts
		WHERE id = $1
	`, accountID)
	return scanParentAccount(row)
}

// UpdateLastLogin records a successful parent login without changing the
// account's other authentication state.
func (r *PostgresRepository) UpdateLastLogin(
	ctx context.Context,
	accountID string,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE parent_accounts
		SET last_login_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
		  AND status = 'active'
	`, accountID)
	if err != nil {
		return fmt.Errorf("update parent last login: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

// CreateSession stores a refresh session.
func (r *PostgresRepository) CreateSession(
	ctx context.Context,
	session *domain.Session,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO auth_sessions (
			id,
			parent_account_id,
			refresh_token_hash,
			expires_at,
			created_at,
			last_used_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		session.ID,
		session.ParentAccountID,
		session.RefreshTokenHash,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastUsedAt,
	)
	if err != nil {
		return fmt.Errorf("insert auth session: %w", err)
	}
	return nil
}

// GetSessionByRefreshTokenHash loads an unrevoked session.
func (r *PostgresRepository) GetSessionByRefreshTokenHash(
	ctx context.Context,
	refreshTokenHash string,
) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			parent_account_id,
			refresh_token_hash,
			expires_at,
			revoked_at,
			created_at,
			last_used_at
		FROM auth_sessions
		WHERE refresh_token_hash = $1
		  AND revoked_at IS NULL
	`, refreshTokenHash)

	var session domain.Session
	err := row.Scan(
		&session.ID,
		&session.ParentAccountID,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
		&session.LastUsedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get auth session: %w", err)
	}
	return &session, nil
}

// RotateSession replaces the refresh hash and expiry after a successful
// refresh. The WHERE clause makes replay of a consumed token fail.
func (r *PostgresRepository) RotateSession(
	ctx context.Context,
	sessionID string,
	refreshTokenHash string,
	expiresAt time.Time,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET refresh_token_hash = $2,
		    expires_at = $3,
		    last_used_at = NOW()
		WHERE id = $1
		  AND revoked_at IS NULL
	`, sessionID, refreshTokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("rotate auth session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

// RevokeSession revokes one session idempotently.
func (r *PostgresRepository) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE id = $1
	`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke auth session: %w", err)
	}
	return nil
}

// RevokeAllSessions revokes every active session for one parent account.
func (r *PostgresRepository) RevokeAllSessions(
	ctx context.Context,
	parentAccountID string,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE parent_account_id = $1
		  AND revoked_at IS NULL
	`, parentAccountID)
	if err != nil {
		return fmt.Errorf("revoke parent sessions: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanParentAccount(row rowScanner) (*domain.ParentAccount, error) {
	var account domain.ParentAccount
	err := row.Scan(
		&account.ID,
		&account.Email,
		&account.Phone,
		&account.PasswordHash,
		&account.DisplayName,
		&account.Status,
		&account.Role,
		&account.GuardianConsentVersion,
		&account.GuardianConsentedAt,
		&account.LastLoginAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan parent account: %w", err)
	}
	return &account, nil
}
