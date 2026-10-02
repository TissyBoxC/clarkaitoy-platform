-- Parent accounts use a verified mobile number as the primary identifier.
-- Email remains optional so guardians can bind it after registration for
-- recovery and alternate sign-in. Existing email-only administrators remain
-- valid through nullable email and the dedicated administrator login flow.
ALTER TABLE parent_accounts
    ALTER COLUMN email DROP NOT NULL;

ALTER TABLE parent_accounts
    ADD COLUMN IF NOT EXISTS guardian_family_name TEXT NOT NULL DEFAULT '';

ALTER TABLE parent_accounts
    ADD COLUMN IF NOT EXISTS child_nickname TEXT;

ALTER TABLE parent_accounts
    ADD COLUMN IF NOT EXISTS child_birthday DATE;

ALTER TABLE parent_accounts
    ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ;

ALTER TABLE parent_accounts
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

-- PostgreSQL permits multiple NULL values under a UNIQUE constraint, so an
-- internal placeholder keeps the new NOT NULL invariant compatible with
-- pre-existing email-only administrator rows without exposing the value.
UPDATE parent_accounts
SET phone = 'legacy:' || id::text
WHERE phone IS NULL OR btrim(phone) = '';

ALTER TABLE parent_accounts
    ALTER COLUMN phone SET NOT NULL;

ALTER TABLE parent_accounts
    DROP CONSTRAINT IF EXISTS parent_accounts_email_not_blank;

ALTER TABLE parent_accounts
    ADD CONSTRAINT parent_accounts_email_not_blank CHECK (
        email IS NULL OR length(btrim(email)) > 3
    );

CREATE INDEX IF NOT EXISTS parent_accounts_child_birthday_idx
    ON parent_accounts(child_birthday)
    WHERE child_birthday IS NOT NULL;
