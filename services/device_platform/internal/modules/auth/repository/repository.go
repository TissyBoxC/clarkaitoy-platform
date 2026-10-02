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
	UpdatePasswordHash(ctx context.Context, accountID string, passwordHash string) error
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
	CreateMFAChallenge(ctx context.Context, challenge *domain.MFAChallenge) error
	GetMFAChallengeByHash(
		ctx context.Context,
		challengeHash string,
	) (*domain.MFAChallenge, error)
	ConsumeMFAChallenge(ctx context.Context, challengeID string) error
	GetTOTPCredential(
		ctx context.Context,
		parentAccountID string,
	) (*domain.TOTPCredential, error)
	UpsertTOTPCredential(ctx context.Context, credential *domain.TOTPCredential) error
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

// UpdatePasswordHash replaces one account's password hash and revokes every
// active session so a reset cannot leave an old login usable.
func (r *PostgresRepository) UpdatePasswordHash(
	ctx context.Context,
	accountID string,
	passwordHash string,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	tag, err := transaction.Exec(ctx, `
		UPDATE parent_accounts
		SET password_hash = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, accountID, passwordHash)
	if err != nil {
		return fmt.Errorf("update parent password hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}

	if _, err := transaction.Exec(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE parent_account_id = $1
		  AND revoked_at IS NULL
	`, accountID); err != nil {
		return fmt.Errorf("revoke sessions after password reset: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit password reset: %w", err)
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

// CreateMFAChallenge stores one hashed, short-lived administrator challenge.
func (r *PostgresRepository) CreateMFAChallenge(
	ctx context.Context,
	challenge *domain.MFAChallenge,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_mfa_challenges (
			id,
			parent_account_id,
			challenge_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		challenge.ID,
		challenge.ParentAccountID,
		challenge.ChallengeHash,
		challenge.ExpiresAt,
		challenge.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert administrator MFA challenge: %w", err)
	}
	return nil
}

// GetMFAChallengeByHash loads one administrator MFA challenge by token hash.
func (r *PostgresRepository) GetMFAChallengeByHash(
	ctx context.Context,
	challengeHash string,
) (*domain.MFAChallenge, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id,
			parent_account_id,
			challenge_hash,
			expires_at,
			consumed_at,
			created_at
		FROM admin_mfa_challenges
		WHERE challenge_hash = $1
	`, challengeHash)

	var challenge domain.MFAChallenge
	err := row.Scan(
		&challenge.ID,
		&challenge.ParentAccountID,
		&challenge.ChallengeHash,
		&challenge.ExpiresAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFAChallengeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get administrator MFA challenge: %w", err)
	}
	return &challenge, nil
}

// ConsumeMFAChallenge atomically prevents replay of an administrator challenge.
func (r *PostgresRepository) ConsumeMFAChallenge(
	ctx context.Context,
	challengeID string,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE admin_mfa_challenges
		SET consumed_at = NOW()
		WHERE id = $1
		  AND consumed_at IS NULL
		  AND expires_at > NOW()
	`, challengeID)
	if err != nil {
		return fmt.Errorf("consume administrator MFA challenge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMFAChallengeConsumed
	}
	return nil
}

// GetTOTPCredential loads the encrypted TOTP enrollment for one administrator.
func (r *PostgresRepository) GetTOTPCredential(
	ctx context.Context,
	parentAccountID string,
) (*domain.TOTPCredential, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			parent_account_id,
			encrypted_secret,
			secret_nonce,
			key_version,
			enabled_at,
			created_at,
			updated_at
		FROM admin_totp_credentials
		WHERE parent_account_id = $1
	`, parentAccountID)

	var credential domain.TOTPCredential
	err := row.Scan(
		&credential.ParentAccountID,
		&credential.EncryptedSecret,
		&credential.SecretNonce,
		&credential.KeyVersion,
		&credential.EnabledAt,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFANotConfigured
	}
	if err != nil {
		return nil, fmt.Errorf("get administrator TOTP credential: %w", err)
	}
	return &credential, nil
}

// UpsertTOTPCredential stores or replaces an administrator TOTP enrollment.
func (r *PostgresRepository) UpsertTOTPCredential(
	ctx context.Context,
	credential *domain.TOTPCredential,
) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_totp_credentials (
			parent_account_id,
			encrypted_secret,
			secret_nonce,
			key_version,
			enabled_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (parent_account_id) DO UPDATE
		SET encrypted_secret = EXCLUDED.encrypted_secret,
		    secret_nonce = EXCLUDED.secret_nonce,
		    key_version = EXCLUDED.key_version,
		    enabled_at = EXCLUDED.enabled_at,
		    updated_at = EXCLUDED.updated_at
	`,
		credential.ParentAccountID,
		credential.EncryptedSecret,
		credential.SecretNonce,
		credential.KeyVersion,
		credential.EnabledAt,
		credential.CreatedAt,
		credential.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert administrator TOTP credential: %w", err)
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
