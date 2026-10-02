CREATE TABLE IF NOT EXISTS parent_accounts (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    phone TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    role TEXT NOT NULL DEFAULT 'parent',
    guardian_consent_version TEXT NOT NULL,
    guardian_consented_at TIMESTAMPTZ NOT NULL,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT parent_accounts_email_not_blank CHECK (length(btrim(email)) > 3),
    CONSTRAINT parent_accounts_role_valid CHECK (role IN ('parent', 'admin')),
    CONSTRAINT parent_accounts_status_valid CHECK (status IN ('active', 'disabled', 'pending_deletion'))
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY,
    parent_account_id UUID NOT NULL REFERENCES parent_accounts(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS auth_sessions_parent_account_id_idx
    ON auth_sessions(parent_account_id);
CREATE INDEX IF NOT EXISTS auth_sessions_expires_at_idx
    ON auth_sessions(expires_at);

CREATE TABLE IF NOT EXISTS admin_mfa_challenges (
    id UUID PRIMARY KEY,
    parent_account_id UUID NOT NULL REFERENCES parent_accounts(id) ON DELETE CASCADE,
    challenge_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS admin_mfa_challenges_account_id_idx
    ON admin_mfa_challenges(parent_account_id);

CREATE TABLE IF NOT EXISTS admin_totp_credentials (
    parent_account_id UUID PRIMARY KEY REFERENCES parent_accounts(id) ON DELETE CASCADE,
    encrypted_secret BYTEA NOT NULL,
    secret_nonce BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    enabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_accounts (
    id UUID PRIMARY KEY,
    parent_account_id UUID NOT NULL UNIQUE REFERENCES parent_accounts(id) ON DELETE CASCADE,
    provider_account_id TEXT NOT NULL UNIQUE,
    provider_account_email TEXT NOT NULL,
    credential_ciphertext BYTEA NOT NULL,
    credential_nonce BYTEA NOT NULL,
    provider_api_key_id BIGINT NOT NULL DEFAULT 0,
    credential_key_version INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'active',
    balance_usd NUMERIC(18, 6) NOT NULL DEFAULT 0,
    concurrency_limit INTEGER NOT NULL DEFAULT 1,
    allowed_models TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ai_accounts_status_valid CHECK (status IN ('active', 'suspended', 'closed')),
    CONSTRAINT ai_accounts_balance_non_negative CHECK (balance_usd >= 0),
    CONSTRAINT ai_accounts_concurrency_positive CHECK (concurrency_limit > 0)
);

CREATE INDEX IF NOT EXISTS ai_accounts_provider_account_id_idx
    ON ai_accounts(provider_account_id);

CREATE TABLE IF NOT EXISTS device_binding_tokens (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS device_binding_tokens_expires_at_idx
    ON device_binding_tokens(expires_at)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS device_binding_tokens_device_id_idx
    ON device_binding_tokens(device_id);

CREATE TABLE IF NOT EXISTS device_bindings (
    id UUID PRIMARY KEY,
    parent_account_id UUID NOT NULL REFERENCES parent_accounts(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL UNIQUE,
    device_name TEXT NOT NULL,
    hardware_model TEXT NOT NULL,
    firmware_version TEXT NOT NULL DEFAULT '',
    capability_set TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    bound_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS device_bindings_parent_account_id_idx
    ON device_bindings(parent_account_id);
