# API v1 Contract

Base path: `/api/v1`
Content type: `application/json`

## MVP Goal

Enable one user to:

1. Maintain reusable foods and recipes.
2. Log meals with one or more items (food or recipe).
3. Track body weight entries.
4. Read daily nutrition totals.

## Error Envelope

All errors use:

```json
{
  "error": {
    "code": "string_code",
    "message": "human readable message",
    "request_id": "request id for log correlation"
  }
}
```

## Existing Endpoints (implemented)

Pagination contract for list endpoints:

- `limit` default `20`, max `100`
- `offset` default `0`, min `0`

## Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

All routes except register/login/refresh/logout and health live/ready require:

- `Authorization: Bearer <jwt>`

## Health

- `GET /health/live` (liveness)
- `GET /health/ready` (readiness with dependency checks)
- `GET /health` (backward-compatible readiness alias)
- Response `200`:

```json
{
  "status": "ok",
  "message": "nutrition api is running"
}
```

## Get User By ID

- `GET /users/{id}`
- `PATCH /users/me`
- Success `200`:

```json
{
  "id": 1,
  "name": "Test User",
  "created_at": "2026-02-17T12:00:00Z",
  "updated_at": "2026-02-17T12:00:00Z"
}
```

User profile fields (optional):

- `sex` (`male|female`)
- `birth_date` (date stored in UTC DB)
- `height_cm`
- `activity_level` (`sedentary|light|moderate|active|very_active`)

- Errors:
  - `400 invalid_user_id`
  - `404 user_not_found`
  - `500 database_error`

## MVP Endpoints (target contract)

## Foods

- `POST /foods`
- `GET /foods`
- `GET /foods/by-barcode/{barcode}`
- `GET /foods/{id}`
- `PATCH /foods/{id}`

Food payload fields:

- `name`
- `barcode` (optional, EAN/UPC digits; scanned lookup key)
- `kcal_per_100g`
- `protein_per_100g`
- `carbs_per_100g`
- `fat_per_100g`

## Recipes

- `POST /recipes`
- `GET /recipes`
- `GET /recipes/{id}`
- `PATCH /recipes/{id}`

Recipe create/update fields:

- `name`
- `yield_weight_g`
- `ingredients`: list of
  - `food_id`
  - `raw_weight_g`

Server computes and stores:

- `kcal_per_100g`
- `protein_per_100g`
- `carbs_per_100g`
- `fat_per_100g`

## Meals

- `POST /meals`
- `POST /meals/{id}/items`
- `GET /meals?date=YYYY-MM-DD`
- `GET /meals/{id}`
- `PATCH /meals/{id}`
- `DELETE /meals/{id}`
- `PATCH /meals/{meal_id}/items/{item_id}`
- `DELETE /meals/{meal_id}/items/{item_id}`

Meal fields:

- `meal_type` (required: `breakfast|lunch|dinner|snack`)
- `eaten_at` (RFC3339 UTC)
- `items` (optional): list of meal item payloads to create atomically with the meal

Meal item fields:

- `weight_g`
- exactly one reference:
  - `food_id`, or
  - `recipe_id`

Server stores per-item nutrition snapshot:

- `kcal_per_100g`
- `protein_per_100g`
- `carbs_per_100g`
- `fat_per_100g`

If `items` is passed to `POST /meals`, meal and items are created in a single database transaction.

## Daily Totals

- `GET /daily-totals?date=YYYY-MM-DD`

Returns aggregate:

```json
{
  "date": "2026-02-17",
  "calories": 2100,
  "protein_g": 140,
  "carbs_g": 220,
  "fat_g": 70
}
```

## User Goals

- `PUT /user-goals` (create/update logged user goals)
- `GET /user-goals`
- `GET /progress/daily?date=YYYY-MM-DD`
- `GET /progress/energy?from=YYYY-MM-DD&to=YYYY-MM-DD`

User goals payload fields:

- `target_kcal`
- `target_protein_g`
- `target_carbs_g`
- `target_fat_g`
- `weight_goal_kg` (optional)
- `activity_level` (optional enum: `sedentary|light|moderate|active|very_active`)

Energy progress returns observed and formula-based TDEE estimate from:

- average intake in period
- body-weight trend in period
- optional profile formula baseline (if profile is complete)

## Body Weight Logs

- `POST /body-weight-logs`
- `GET /body-weight-logs?from=YYYY-MM-DD&to=YYYY-MM-DD`
- `GET /body-weight-logs/latest`

Body weight payload fields:

- `weight_kg`
- `logged_at` (RFC3339 UTC)

## API Rules

1. Use UTC timestamps in RFC3339.
2. List endpoints must support pagination (`limit`, `offset`) by default.
3. Validate input and return stable error codes.
4. Keep backward compatibility inside v1.
5. `meal_items` must satisfy `food_id XOR recipe_id`.
6. Nutrition values and all weight fields must be positive/non-negative as defined in domain invariants.
