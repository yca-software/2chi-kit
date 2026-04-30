.PHONY: help test migrate-create swagger

# Repository packages share one Postgres (testcontainers reuse or TWOCHI_KIT_GO_API_INTEGRATION_TEST_DSN); they must not run in parallel.
REPO_TEST_PKGS := ./internals/repositories/...
NON_REPO_TEST_PKGS := $(shell go list ./... | grep -v '/internals/repositories/')

# Default target
help:
	@echo "Example API - Available commands:"
	@echo ""
	@echo "  make test             - Run tests (-timeout 10m). Repositories run with -p 1 (shared DB)."
	@echo "  make test-integration - Same as repository slice only (shared DB, -p 1)"
	@echo "  Env TWOCHI_KIT_GO_API_INTEGRATION_TEST_DSN - Use existing Postgres/PostGIS (skip testcontainers)"
	@echo "  Env TWOCHI_KIT_GO_API_INTEGRATION_TEST_NO_CONTAINER_REUSE=1 - New container per process (slow)"
	@echo "  make swagger          - Generate Swagger docs (docs/docs.go, docs/swagger.json, docs/swagger.yaml)"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-create NAME=name - Create new migration"

# Swagger: generate full API docs (paths from handlers). Requires: go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g main.go -d "cmd/api,internals/transport/http,internals/models,internals/services" -o docs --outputTypes json,yaml --parseInternal

# Tests: non-repository packages in parallel; repositories serialized (-p 1) so one shared integration DB is safe.
test:
	go test $(NON_REPO_TEST_PKGS) -count=1 -timeout 25m
	go test $(REPO_TEST_PKGS) -count=1 -p 1 -timeout 25m

test-short:
	go test ./... -count=1 -short

# Integration tests (repositories): shared container reuse or TWOCHI_KIT_GO_API_INTEGRATION_TEST_DSN; -p 1 only.
test-integration:
	go test $(REPO_TEST_PKGS) -count=1 -p 1 -timeout 25m

# Migrations - requires DATABASE_POSTGRES_DSN
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME required. Usage: make migrate-create NAME=add_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)
