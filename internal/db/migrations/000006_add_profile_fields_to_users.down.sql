ALTER TABLE users DROP CONSTRAINT IF EXISTS users_activity_level_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_height_cm_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_sex_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS activity_level,
    DROP COLUMN IF EXISTS height_cm,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS sex;
