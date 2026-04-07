-- 002_google_oauth.sql
-- Allow users to exist without a local password (Google-only sign-in).
-- Add avatar_url populated from Google userinfo.

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
