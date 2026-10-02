CREATE TABLE IF NOT EXISTS device_credentials (
    device_id TEXT PRIMARY KEY,
    hardware_model TEXT NOT NULL DEFAULT '',
    firmware_version TEXT NOT NULL DEFAULT '',
    capability_set TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    public_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_authenticated_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT device_credentials_status_valid CHECK (status IN ('active', 'disabled', 'revoked')),
    CONSTRAINT device_credentials_public_key_not_blank CHECK (length(btrim(public_key)) > 0)
);

CREATE INDEX IF NOT EXISTS device_credentials_status_idx
    ON device_credentials(status);

CREATE TABLE IF NOT EXISTS device_registration_tokens (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS device_registration_tokens_device_id_idx
    ON device_registration_tokens(device_id);

CREATE TABLE IF NOT EXISTS device_challenges (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES device_credentials(device_id) ON DELETE CASCADE,
    nonce_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS device_challenges_device_id_idx
    ON device_challenges(device_id);

CREATE TABLE IF NOT EXISTS device_sessions (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES device_credentials(device_id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS device_sessions_device_id_idx
    ON device_sessions(device_id);
