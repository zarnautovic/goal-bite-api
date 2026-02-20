# Nutrition

Go + Postgres playground using VS Code Dev Containers.

## Included

- Go 1.22 dev container
- Postgres 16 service (`db`)
- `DATABASE_URL` and standard `PG*` env vars preconfigured inside container

## Start

1. Open this folder in VS Code.
2. Run: `Dev Containers: Reopen in Container`.
3. Inside container, run:
   - one-time setup: `make create-test-db && make migrate-up && make seed`
   - start API: `make run`

## Local Env File

- Copy `.env.example` to `.env` and adjust values when running outside the container.
- JWT rotation envs:
  - `JWT_ACTIVE_KID` (default `v1`)
  - `JWT_KEYS` format: `kid:secret,kid2:secret2`
  - if `JWT_KEYS` is empty, app uses `JWT_SECRET` for `JWT_ACTIVE_KID`
- Auth login hardening envs:
  - `AUTH_LOGIN_MAX_ATTEMPTS` (default `5`)
  - `AUTH_LOGIN_WINDOW_MINUTES` (default `10`)
  - `AUTH_LOGIN_LOCKOUT_MINUTES` (default `15`)

## API

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`
- `GET /api/v1/auth/me`
- `GET /api/v1/health`
- `GET /api/v1/users/{id}`
- `PATCH /api/v1/users/me`
- `POST /api/v1/foods`
- `GET /api/v1/foods`
- `GET /api/v1/foods/by-barcode/{barcode}`
- `GET /api/v1/foods/{id}`
- `PATCH /api/v1/foods/{id}`
- `POST /api/v1/recipes`
- `GET /api/v1/recipes`
- `GET /api/v1/recipes/{id}`
- `PATCH /api/v1/recipes/{id}`
- `POST /api/v1/meals`
- `GET /api/v1/meals?date=YYYY-MM-DD&limit=20&offset=0`
- `GET /api/v1/meals/{id}`
- `PATCH /api/v1/meals/{id}`
- `DELETE /api/v1/meals/{id}`
- `POST /api/v1/meals/{id}/items`
- `PATCH /api/v1/meals/{meal_id}/items/{item_id}`
- `DELETE /api/v1/meals/{meal_id}/items/{item_id}`
- `GET /api/v1/daily-totals?date=YYYY-MM-DD`
- `PUT /api/v1/user-goals`
- `GET /api/v1/user-goals`
- `GET /api/v1/progress/daily?date=YYYY-MM-DD`
- `GET /api/v1/progress/energy?from=YYYY-MM-DD&to=YYYY-MM-DD`
- `POST /api/v1/body-weight-logs`
- `GET /api/v1/body-weight-logs?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=20&offset=0`
- `GET /api/v1/body-weight-logs/latest`
- Swagger UI: `GET /swagger/index.html`

All routes except `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, `GET /api/v1/health/live`, and `GET /api/v1/health/ready` require:
- `Authorization: Bearer <jwt>`

## Planning Docs

- Product scope: `docs/product-scope.md`
- Domain model: `docs/domain-model.md`
- API v1 contract: `docs/api-v1.md`
- OpenAPI error code catalog: `docs/openapi-error-codes.md`
- MVP backlog: `docs/mvp-backlog.md`
- Architecture decisions: `docs/adr/0001-foundation-decisions.md`
- Logging policy: `docs/logging.md`

## Migrations

- Up: `go run ./cmd/migrate -direction up`
- Down one step: `go run ./cmd/migrate -direction down -steps 1`
- Migrations are manual and are not executed by API startup.

## Dev Workflow

- `make help` to list commands
- `make compose-up` to start app + db containers
- `make shell` to open a shell in the app container
- `make psql` to open Postgres client in the app container
- `make test` to run tests
- `make test-integration` to run Postgres-backed integration/e2e tests (uses `TEST_DATABASE_URL`)
  - uses `TEST_DATABASE_URL` (defaults to `nutrition_test`) so app data stays safe
- `make lint` to run `go vet`
- `make build` to compile API
- `make swagger-install` to install `swag` CLI
- `make swagger` to generate OpenAPI docs in `docs/swagger`
- `make logs` to stream container logs
- `make compose-down` to stop containers
- `make dev` to run with hot reload via `air` (if installed)
- `make create-test-db` to create `nutrition_test` database
- `make drop-test-db` to remove `nutrition_test` database

## CI

- GitHub Actions workflow: `.github/workflows/ci.yml`
- `unit` job: `go vet`, `go test`, `go build`, swagger generation check
- `integration` job: boots Postgres service and runs `make test-integration`

## Bruno API Client

- Collection path: `bruno/`
- Open/import the `bruno` folder in Bruno.
- Select environment: `local`.
- Suggested auth/user workflow:
  - `bruno/auth/register`
  - `bruno/auth/login` (stores `jwt` + `refreshToken`)
  - `bruno/auth/me` (verify current session user)
  - `bruno/users/update_me` (update profile fields)
  - `bruno/auth/refresh` (rotate tokens)
  - `bruno/auth/logout` (revoke refresh session)
- Run requests in:
  - `bruno/auth/`
  - `bruno/health/`
  - `bruno/users/`
  - `bruno/foods/`
  - `bruno/recipes/`
  - `bruno/meals/`
  - `bruno/user-goals/`
  - `bruno/progress/`
  - `bruno/body-weight-logs/`
