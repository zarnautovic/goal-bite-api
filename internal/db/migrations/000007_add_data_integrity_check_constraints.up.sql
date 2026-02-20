ALTER TABLE foods
    ADD CONSTRAINT foods_kcal_per_100g_max_check CHECK (kcal_per_100g <= 900),
    ADD CONSTRAINT foods_protein_per_100g_max_check CHECK (protein_per_100g <= 100),
    ADD CONSTRAINT foods_carbs_per_100g_max_check CHECK (carbs_per_100g <= 100),
    ADD CONSTRAINT foods_fat_per_100g_max_check CHECK (fat_per_100g <= 100);

ALTER TABLE recipes
    ADD CONSTRAINT recipes_yield_weight_g_max_check CHECK (yield_weight_g <= 100000),
    ADD CONSTRAINT recipes_kcal_per_100g_max_check CHECK (kcal_per_100g <= 900),
    ADD CONSTRAINT recipes_protein_per_100g_max_check CHECK (protein_per_100g <= 100),
    ADD CONSTRAINT recipes_carbs_per_100g_max_check CHECK (carbs_per_100g <= 100),
    ADD CONSTRAINT recipes_fat_per_100g_max_check CHECK (fat_per_100g <= 100);

ALTER TABLE recipe_ingredients
    ADD CONSTRAINT recipe_ingredients_raw_weight_g_max_check CHECK (raw_weight_g <= 100000);

ALTER TABLE meal_items
    ADD CONSTRAINT meal_items_weight_g_max_check CHECK (weight_g <= 100000),
    ADD CONSTRAINT meal_items_kcal_per_100g_max_check CHECK (kcal_per_100g <= 900),
    ADD CONSTRAINT meal_items_protein_per_100g_max_check CHECK (protein_per_100g <= 100),
    ADD CONSTRAINT meal_items_carbs_per_100g_max_check CHECK (carbs_per_100g <= 100),
    ADD CONSTRAINT meal_items_fat_per_100g_max_check CHECK (fat_per_100g <= 100);

ALTER TABLE body_weight_logs
    ADD CONSTRAINT body_weight_logs_weight_kg_max_check CHECK (weight_kg <= 500);

ALTER TABLE users
    ADD CONSTRAINT users_height_cm_max_check CHECK (height_cm <= 300 OR height_cm IS NULL);
