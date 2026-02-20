# Domain Model

## MVP Scope

MVP supports:

1. Reusable foods with nutrition values per 100g.
2. Reusable recipes built from raw ingredient foods.
3. Meal logging with one or more items.
4. Meal item can reference either a food or a recipe.
5. Body weight logging for future BMR/TDEE calculations.

## Entities

## User

- `id` (bigint, PK)
- `name` (text, required)
- `created_at` / `updated_at` (timestamptz)

Future expansion:
- profile/targets fields (`email`, `timezone`, calorie/macro targets, activity level).

## Food

Reusable standalone food entry.

- `id` (bigint, PK)
- `name` (text, required)
- `kcal_per_100g` (numeric, required)
- `protein_per_100g` (numeric, required)
- `carbs_per_100g` (numeric, required)
- `fat_per_100g` (numeric, required)
- `created_at` / `updated_at` (timestamptz)

Notes:
- Allows manual food creation (for example, user can directly create `goulash` as a food).

## Recipe

Reusable recipe entry derived from raw ingredients.

- `id` (bigint, PK)
- `name` (text, required)
- `yield_weight_g` (numeric, required) // final cooked total weight
- `kcal_per_100g` (numeric, required, computed)
- `protein_per_100g` (numeric, required, computed)
- `carbs_per_100g` (numeric, required, computed)
- `fat_per_100g` (numeric, required, computed)
- `created_at` / `updated_at` (timestamptz)

## RecipeIngredient

Raw ingredients used to build a recipe.

- `id` (bigint, PK)
- `recipe_id` (FK -> recipes.id, required)
- `food_id` (FK -> foods.id, required)
- `raw_weight_g` (numeric, required)
- `position` (int, optional)
- `created_at` / `updated_at` (timestamptz)

Recipe nutrition computation rule:

1. User logs ingredients as raw weights.
2. Total recipe nutrients = sum of ingredient nutrients.
3. User provides final cooked `yield_weight_g`.
4. Per-100g values = total nutrients / (`yield_weight_g` / 100).

## Meal

- `id` (bigint, PK)
- `user_id` (FK -> users.id, required)
- `meal_type` (text, required; e.g. breakfast/lunch/dinner/snack)
- `eaten_at` (timestamptz, required)
- `created_at` / `updated_at` (timestamptz)

## MealItem

One item in a meal, from either food or recipe.

- `id` (bigint, PK)
- `meal_id` (FK -> meals.id, required)
- `food_id` (FK -> foods.id, nullable)
- `recipe_id` (FK -> recipes.id, nullable)
- `weight_g` (numeric, required)
- `kcal_per_100g` (numeric, required, snapshot)
- `protein_per_100g` (numeric, required, snapshot)
- `carbs_per_100g` (numeric, required, snapshot)
- `fat_per_100g` (numeric, required, snapshot)
- `created_at` / `updated_at` (timestamptz)

Constraint:

- Exactly one of `food_id` or `recipe_id` must be set.

Snapshot rule:

- Nutrition values are copied at meal log time so historical logs stay stable if food/recipe definitions change later.

## BodyWeightLog

- `id` (bigint, PK)
- `user_id` (FK -> users.id, required)
- `weight_kg` (numeric, required)
- `logged_at` (timestamptz, required)
- `created_at` / `updated_at` (timestamptz)

Purpose:

- Store data needed for future basal metabolism, TDEE, and trend calculations.

## Relationships

1. `recipes 1..n recipe_ingredients`
2. `users 1..n meals`
3. `meals 1..n meal_items`
4. `foods 1..n meal_items` (optional reference)
5. `recipes 1..n meal_items` (optional reference)
6. `users 1..n body_weight_logs`

## Ownership Rules

1. User can access only their own meals and weight logs.
2. Foods and recipes are global and reusable by all users in MVP.
3. Meal items cannot exist without a parent meal.
4. Recipe ingredients cannot exist without a parent recipe.

## Invariants

1. Nutrition values are non-negative.
2. `weight_g`, `raw_weight_g`, and `yield_weight_g` are positive.
3. `meal_type` must be one of the allowed values.
4. Exactly one reference in `meal_items`: `food_id XOR recipe_id`.
5. `updated_at` changes on modification.
