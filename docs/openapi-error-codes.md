# OpenAPI Error Code Catalog

This catalog defines stable API error codes returned in:

```json
{
  "error": {
    "code": "some_code",
    "message": "human readable message"
  }
}
```

## Common

- `database_error`: unexpected persistence/runtime failure.
- `route_not_found`: route does not exist.
- `method_not_allowed`: HTTP method is not allowed on route.
- `invalid_request_body`: malformed JSON or unknown fields.
- `invalid_pagination`: invalid `limit`/`offset`.
- `unauthorized`: missing/invalid bearer token.
- `forbidden`: authenticated user does not own resource.
- `rate_limited`: too many requests in the current window.
- `service_unavailable`: service dependency is not ready.

## Auth

- `invalid_register_payload`
- `invalid_password_policy`
- `invalid_login_payload`
- `invalid_refresh_payload`
- `invalid_logout_payload`
- `invalid_credentials`
- `too_many_login_attempts`
- `invalid_refresh_token`
- `email_already_exists`

## Users

- `invalid_user_id`
- `invalid_user_payload`
- `user_not_found`

## Foods

- `invalid_food_id`
- `invalid_food_barcode`
- `invalid_food_payload`
- `food_not_found`
- `food_barcode_not_found`
- `food_barcode_already_exists`

## Recipes

- `invalid_recipe_id`
- `invalid_recipe_payload`
- `recipe_not_found`
- `ingredient_food_not_found`

## Meals

- `invalid_meal_id`
- `invalid_meal_item_id`
- `invalid_meal_payload`
- `invalid_meal_query`
- `invalid_meal_item_payload`
- `meal_not_found`
- `meal_item_not_found`
- `food_not_found`
- `recipe_not_found`
- `invalid_daily_totals_query`

## Body Weight Logs

- `invalid_body_weight_payload`
- `invalid_body_weight_query`
- `body_weight_log_not_found`

## User Goals

- `invalid_user_goals_payload`
- `user_goals_not_found`
- `invalid_daily_progress_query`

## Progress

- `invalid_energy_progress_query`
- `insufficient_weight_data`
- `insufficient_intake_data`

## Notes

- Error `message` is human-readable and may evolve.
- Error `code` is stable and should be used for client-side branching.
