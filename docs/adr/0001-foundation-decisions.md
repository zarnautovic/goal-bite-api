# ADR 0001: Foundation Decisions

## Status

Accepted

## Date

2026-02-17

## Context

We need a pragmatic backend foundation that is easy to iterate on while preserving maintainability.

## Decisions

1. Language/runtime: Go (standard toolchain).
2. HTTP routing: `chi` (`github.com/go-chi/chi/v5`).
3. ORM: GORM.
4. Migrations: `golang-migrate` with SQL files.
5. API prefix: `/api/v1`.
6. Error shape: standard envelope with `code` and `message`.
7. Architecture: layered (`handlers` -> `service` -> `repository`).
8. Development: Dev Container + Postgres via Docker Compose.

## Consequences

Positive:

1. Fast development with clear layering.
2. SQL schema history is explicit and reviewable.
3. API versioning is established early.

Tradeoffs:

1. GORM flexibility can hide SQL details if not reviewed.
2. Migration discipline is required (no schema drift outside migrations).
3. Additional layers add boilerplate but improve long-term maintainability.
4. Migrations and seed are executed manually (`cmd/migrate`, `cmd/seed`), not during API startup.

## Follow-up ADRs

1. Auth strategy (JWT/session).
2. Input validation library choice.
3. Logging/tracing/metrics standard.
4. Pagination standard (`offset` vs cursor).
