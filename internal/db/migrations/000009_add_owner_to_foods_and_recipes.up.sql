ALTER TABLE foods ADD COLUMN IF NOT EXISTS user_id BIGINT;
ALTER TABLE recipes ADD COLUMN IF NOT EXISTS user_id BIGINT;

UPDATE foods
SET user_id = (
    SELECT id
    FROM users
    ORDER BY id ASC
    LIMIT 1
)
WHERE user_id IS NULL;

UPDATE recipes
SET user_id = (
    SELECT id
    FROM users
    ORDER BY id ASC
    LIMIT 1
)
WHERE user_id IS NULL;

ALTER TABLE foods
    ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE recipes
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE foods
    ADD CONSTRAINT fk_foods_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

ALTER TABLE recipes
    ADD CONSTRAINT fk_recipes_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_foods_user_id ON foods(user_id);
CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipes(user_id);
