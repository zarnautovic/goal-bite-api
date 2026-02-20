CREATE TYPE meal_type_enum AS ENUM ('breakfast', 'lunch', 'dinner', 'snack');

CREATE TABLE IF NOT EXISTS foods (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    kcal_per_100g NUMERIC(12,4) NOT NULL CHECK (kcal_per_100g >= 0),
    protein_per_100g NUMERIC(12,4) NOT NULL CHECK (protein_per_100g >= 0),
    carbs_per_100g NUMERIC(12,4) NOT NULL CHECK (carbs_per_100g >= 0),
    fat_per_100g NUMERIC(12,4) NOT NULL CHECK (fat_per_100g >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recipes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    yield_weight_g NUMERIC(12,4) NOT NULL CHECK (yield_weight_g > 0),
    kcal_per_100g NUMERIC(12,4) NOT NULL CHECK (kcal_per_100g >= 0),
    protein_per_100g NUMERIC(12,4) NOT NULL CHECK (protein_per_100g >= 0),
    carbs_per_100g NUMERIC(12,4) NOT NULL CHECK (carbs_per_100g >= 0),
    fat_per_100g NUMERIC(12,4) NOT NULL CHECK (fat_per_100g >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id BIGSERIAL PRIMARY KEY,
    recipe_id BIGINT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    food_id BIGINT NOT NULL REFERENCES foods(id) ON DELETE RESTRICT,
    raw_weight_g NUMERIC(12,4) NOT NULL CHECK (raw_weight_g > 0),
    position INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    meal_type meal_type_enum NOT NULL,
    eaten_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meal_items (
    id BIGSERIAL PRIMARY KEY,
    meal_id BIGINT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    food_id BIGINT REFERENCES foods(id) ON DELETE RESTRICT,
    recipe_id BIGINT REFERENCES recipes(id) ON DELETE RESTRICT,
    weight_g NUMERIC(12,4) NOT NULL CHECK (weight_g > 0),
    kcal_per_100g NUMERIC(12,4) NOT NULL CHECK (kcal_per_100g >= 0),
    protein_per_100g NUMERIC(12,4) NOT NULL CHECK (protein_per_100g >= 0),
    carbs_per_100g NUMERIC(12,4) NOT NULL CHECK (carbs_per_100g >= 0),
    fat_per_100g NUMERIC(12,4) NOT NULL CHECK (fat_per_100g >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT meal_items_single_source CHECK (
        (food_id IS NOT NULL AND recipe_id IS NULL)
        OR
        (food_id IS NULL AND recipe_id IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS body_weight_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight_kg NUMERIC(12,4) NOT NULL CHECK (weight_kg > 0),
    logged_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_foods_name ON foods(name);
CREATE INDEX IF NOT EXISTS idx_recipes_name ON recipes(name);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);
CREATE INDEX IF NOT EXISTS idx_meals_user_id_eaten_at ON meals(user_id, eaten_at);
CREATE INDEX IF NOT EXISTS idx_meal_items_meal_id ON meal_items(meal_id);
CREATE INDEX IF NOT EXISTS idx_body_weight_logs_user_id_logged_at ON body_weight_logs(user_id, logged_at);
