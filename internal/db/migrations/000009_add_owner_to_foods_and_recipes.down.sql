DROP INDEX IF EXISTS idx_recipes_user_id;
DROP INDEX IF EXISTS idx_foods_user_id;

ALTER TABLE recipes DROP CONSTRAINT IF EXISTS fk_recipes_user_id;
ALTER TABLE foods DROP CONSTRAINT IF EXISTS fk_foods_user_id;

ALTER TABLE recipes DROP COLUMN IF EXISTS user_id;
ALTER TABLE foods DROP COLUMN IF EXISTS user_id;
