ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_height_cm_max_check;

ALTER TABLE body_weight_logs
    DROP CONSTRAINT IF EXISTS body_weight_logs_weight_kg_max_check;

ALTER TABLE meal_items
    DROP CONSTRAINT IF EXISTS meal_items_fat_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS meal_items_carbs_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS meal_items_protein_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS meal_items_kcal_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS meal_items_weight_g_max_check;

ALTER TABLE recipe_ingredients
    DROP CONSTRAINT IF EXISTS recipe_ingredients_raw_weight_g_max_check;

ALTER TABLE recipes
    DROP CONSTRAINT IF EXISTS recipes_fat_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS recipes_carbs_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS recipes_protein_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS recipes_kcal_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS recipes_yield_weight_g_max_check;

ALTER TABLE foods
    DROP CONSTRAINT IF EXISTS foods_fat_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS foods_carbs_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS foods_protein_per_100g_max_check,
    DROP CONSTRAINT IF EXISTS foods_kcal_per_100g_max_check;
