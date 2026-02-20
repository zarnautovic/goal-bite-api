# MVP Backlog

This backlog is ordered for safe, incremental delivery.
Each phase should keep the API runnable and testable.

## Phase 0: Foundation Check

1. Confirm current baseline works (`make run`, `make test`, `make lint`).
2. Keep migrations/seed manual only.
3. Keep request logging middleware active.

## Phase 1: Schema Expansion (Migrations)

Goal: introduce core MVP tables and constraints.

Tasks:

1. Add migration: `foods`.
2. Add migration: `recipes`.
3. Add migration: `recipe_ingredients`.
4. Add migration: `meals` (with required `meal_type`).
5. Add migration: `meal_items`.
6. Add migration: `body_weight_logs`.
7. Add DB constraints:
   - positive weights (`raw_weight_g`, `yield_weight_g`, `weight_g`, `weight_kg`)
   - non-negative nutrition values
   - `meal_items` check: exactly one of `food_id` or `recipe_id`
8. Add indexes:
   - `meals(user_id, eaten_at)`
   - `meal_items(meal_id)`
   - `recipe_ingredients(recipe_id)`
   - `body_weight_logs(user_id, logged_at)`

Definition of done:

1. `make migrate-up` creates full schema.
2. `make migrate-down` rolls back cleanly.

## Phase 2: Domain + Repository Layer

Goal: represent new entities in code and persist them.

Tasks:

1. Add domain models: food, recipe, recipe_ingredient, meal, meal_item, body_weight_log.
2. Add repositories for each aggregate.
3. Add repository integration tests (focused, real Postgres).

Definition of done:

1. Create/read operations work in repository tests.
2. Constraints are enforced at DB level.

## Phase 3: Service Layer Rules

Goal: centralize business logic and calculations.

Tasks:

1. Food service: create/list/get/update.
2. Recipe service:
   - validate ingredient list
   - compute per-100g from raw ingredients + `yield_weight_g`
3. Meal service:
   - create meal with required `meal_type`
   - add item by food or recipe
   - store nutrition snapshot in `meal_items`
4. Body weight service: add/list/latest.
5. Daily totals service: aggregate by date.

Definition of done:

1. Unit tests cover calculations and validation logic.
2. Service returns stable domain errors.

## Phase 4: HTTP API Endpoints

Goal: expose MVP contract at `/api/v1`.

Tasks:

1. Users:
   - `POST /users`
   - `PATCH /users/{id}`
2. Foods:
   - `POST /foods`
   - `GET /foods`
   - `GET /foods/{id}`
   - `PATCH /foods/{id}`
3. Recipes:
   - `POST /recipes`
   - `GET /recipes`
   - `GET /recipes/{id}`
   - `PATCH /recipes/{id}`
4. Meals:
   - `POST /meals`
   - `POST /meals/{id}/items`
   - `GET /meals?date=YYYY-MM-DD`
   - `GET /meals/{id}`
5. Daily totals:
   - `GET /daily-totals?date=YYYY-MM-DD`
6. Body weight logs:
   - `POST /body-weight-logs`
   - `GET /body-weight-logs?from=YYYY-MM-DD&to=YYYY-MM-DD`
   - `GET /body-weight-logs/latest`

Definition of done:

1. Handlers have request validation.
2. Errors follow the shared envelope.
3. Handler tests cover success + validation + not found + internal errors.

## Phase 5: Hardening

Goal: reduce regressions and prepare for growth.

Tasks:

1. Add end-to-end smoke tests for core flows.
2. Add pagination defaults for list endpoints.
3. Add OpenAPI spec draft (`docs/openapi.yaml`) matching `docs/api-v1.md`.
4. Add feature flags/config for future auth enablement.

Definition of done:

1. `make test` passes all layers.
2. API contract docs and behavior match.

## Suggested Implementation Order (Vertical Slices)

1. Foods (schema -> repo -> service -> handler -> tests)
2. Recipes (+ ingredient math + tests)
3. Meals + meal_items (+ snapshot logic)
4. Daily totals aggregation
5. Body weight logs

This order gives reusable primitives before higher-level aggregation.
