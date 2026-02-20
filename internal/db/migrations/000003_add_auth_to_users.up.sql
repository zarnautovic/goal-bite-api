ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE users
SET email = CONCAT('legacy-user-', id, '@local.invalid')
WHERE email IS NULL;

UPDATE users
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE password_hash IS NULL;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email);
