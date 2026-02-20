.PHONY: help run build test test-integration lint migrate-up migrate-down seed fmt tidy compose-up compose-down compose-restart logs shell psql dev swagger swagger-install create-test-db drop-test-db

GO ?= go
COMPOSE ?= docker compose -f .devcontainer/docker-compose.yml
APP_SERVICE ?= app
DB_SERVICE ?= db

help:
	@echo "Available targets:"
	@echo "  make run           - start API server"
	@echo "  make build         - build API binary"
	@echo "  make test          - run all tests"
	@echo "  make test-integration - run integration/e2e tests (requires TEST_DATABASE_URL)"
	@echo "  make lint          - run static checks (go vet)"
	@echo "  make migrate-up    - apply all up migrations"
	@echo "  make migrate-down  - rollback one migration step"
	@echo "  make seed          - seed local test data once"
	@echo "  make fmt           - format Go code"
	@echo "  make tidy          - tidy Go modules"
	@echo "  make compose-up    - start app + db containers"
	@echo "  make compose-down  - stop containers"
	@echo "  make compose-restart - restart containers"
	@echo "  make logs          - follow docker compose logs"
	@echo "  make shell         - open shell in app container"
	@echo "  make psql          - open psql in app container"
	@echo "  make dev           - run with air (if installed)"
	@echo "  make swagger-install - install swag CLI"
	@echo "  make swagger       - generate OpenAPI docs"
	@echo "  make create-test-db - create nutrition_test database in postgres container"
	@echo "  make drop-test-db   - drop nutrition_test database in postgres container"

run:
	$(GO) run ./cmd/api

build:
	$(GO) build ./cmd/api

test:
	$(GO) test ./...

test-integration:
	@if [ -z "$$TEST_DATABASE_URL" ]; then \
		if [ -f /.dockerenv ]; then \
			export TEST_DATABASE_URL=postgres://postgres:postgres@db:5432/nutrition_test?sslmode=disable; \
		else \
			export TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/nutrition_test?sslmode=disable; \
		fi; \
	fi; \
	$(GO) test -tags=integration ./internal/e2e -v

lint:
	$(GO) vet ./...

migrate-up:
	$(GO) run ./cmd/migrate -direction up

migrate-down:
	$(GO) run ./cmd/migrate -direction down -steps 1

seed:
	$(GO) run ./cmd/seed

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

compose-up:
	$(COMPOSE) up -d

compose-down:
	$(COMPOSE) down

compose-restart:
	$(COMPOSE) up -d --force-recreate

logs:
	$(COMPOSE) logs -f

shell:
	$(COMPOSE) exec $(APP_SERVICE) bash

psql:
	$(COMPOSE) exec $(DB_SERVICE) psql -U postgres -d nutrition

create-test-db:
	$(COMPOSE) exec $(DB_SERVICE) bash -lc "psql -U postgres -d postgres -tc \"SELECT 1 FROM pg_database WHERE datname='nutrition_test'\" | grep -q 1 || psql -U postgres -d postgres -c \"CREATE DATABASE nutrition_test;\""

drop-test-db:
	$(COMPOSE) exec $(DB_SERVICE) psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS nutrition_test;"

dev:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air is not installed. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

swagger-install:
	$(GO) install github.com/swaggo/swag/cmd/swag@latest

swagger:
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs/swagger; \
	else \
		$(GO) run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs/swagger; \
	fi
