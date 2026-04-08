-- Reverse: 002_google_oauth.up.sql
-- Note: setting password_hash NOT NULL will fail if any Google-only users
-- exist. This is the correct behavior for a real downgrade — drop the
-- Google-only users first if you need to apply this.
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
