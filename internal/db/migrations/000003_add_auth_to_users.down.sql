DROP INDEX IF EXISTS idx_users_email_unique;

ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
ALTER TABLE users DROP COLUMN IF EXISTS email;
