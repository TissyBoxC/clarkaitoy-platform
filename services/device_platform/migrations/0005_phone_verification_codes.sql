-- SMS verification is optional during local development. Persisting only the
-- code hash keeps the interface ready for a production SMS provider without
-- storing plaintext codes.
CREATE TABLE IF NOT EXISTS phone_verification_codes (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL,
    purpose TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT phone_verification_codes_purpose_valid CHECK (
        purpose IN ('register', 'login', 'bind_email')
    )
);

CREATE INDEX IF NOT EXISTS phone_verification_codes_phone_idx
    ON phone_verification_codes(phone, purpose, created_at DESC);

CREATE INDEX IF NOT EXISTS phone_verification_codes_expires_at_idx
    ON phone_verification_codes(expires_at)
    WHERE consumed_at IS NULL;
