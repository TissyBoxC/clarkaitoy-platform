-- Split the provider-approved model pool from the guardian's selection.
--
-- available_models is the platform-approved menu shown to guardians.
-- selected_models is what the guardian enabled; an empty selection means
-- "all available models". allowed_models keeps the effective allowlist that
-- was pushed to the provider so the two sides stay auditable.
ALTER TABLE ai_accounts
    ADD COLUMN IF NOT EXISTS available_models TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

ALTER TABLE ai_accounts
    ADD COLUMN IF NOT EXISTS selected_models TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

-- Existing rows stored the guardian selection in allowed_models. Treat it as
-- both the available pool and the current selection so no access is widened.
UPDATE ai_accounts
SET available_models = allowed_models,
    selected_models = allowed_models
WHERE available_models = ARRAY[]::TEXT[]
  AND allowed_models <> ARRAY[]::TEXT[];
