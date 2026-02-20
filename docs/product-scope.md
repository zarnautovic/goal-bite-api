# Product Scope

## Product Goal

Nutrition API helps users track food intake and monitor daily nutrition targets.

## Primary Users

- Individual users who want to log meals and see calories/macros.
- Future: coaches/admin users (out of current v1 scope).

## Problem Statement

Users need a simple way to record what they eat and understand nutrition totals by day.

## v1 Scope

1. User profile basics.
2. Food catalog (search + read).
3. Meal logging for a selected date.
4. Daily nutrition totals.

## Out of Scope (v1)

- Social/sharing features.
- Recommendation engine.
- Barcode scanning.
- Wearable integrations.

## Success Criteria (v1)

1. User can create profile and retrieve it.
2. User can add meal items for a date.
3. User can retrieve daily totals for calories/protein/carbs/fat.
4. API errors are consistent and actionable.

## Milestones

1. Foundation (done/in progress): architecture, migrations, health, get user.
2. User flows: create/update user profile.
3. Food flows: add/list/search foods.
4. Logging flows: create meals, add meal items, compute daily totals.
5. Hardening: tests, auth baseline, observability.
