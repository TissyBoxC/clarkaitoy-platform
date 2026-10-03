-- Release versions are globally unique across kinds and platforms.
--
-- The management API addresses one publication by version, so the previous
-- composite key allowed two rows to share a path that had no unique target.

ALTER TABLE platform_releases
    DROP CONSTRAINT IF EXISTS platform_releases_version_unique;

ALTER TABLE platform_releases
    ADD CONSTRAINT platform_releases_version_unique UNIQUE (version);
