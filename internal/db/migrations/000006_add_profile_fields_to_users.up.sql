ALTER TABLE users
    ADD COLUMN IF NOT EXISTS sex TEXT,
    ADD COLUMN IF NOT EXISTS birth_date DATE,
    ADD COLUMN IF NOT EXISTS height_cm NUMERIC(6,2),
    ADD COLUMN IF NOT EXISTS activity_level TEXT;

ALTER TABLE users
    ADD CONSTRAINT users_sex_check CHECK (sex IN ('male', 'female') OR sex IS NULL);

ALTER TABLE users
    ADD CONSTRAINT users_height_cm_check CHECK (height_cm > 0 OR height_cm IS NULL);

ALTER TABLE users
    ADD CONSTRAINT users_activity_level_check CHECK (
        activity_level IN ('sedentary', 'light', 'moderate', 'active', 'very_active')
        OR activity_level IS NULL
    );
