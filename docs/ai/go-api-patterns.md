# Go API patterns (AI)

## Purpose

HTTP API built on Echo, Postgres, Redis, RabbitMQ jobs, shared `go-common`, SQL migrations, Swag docs.

## Folder structure (canonical)

Relative to `apps/template/go-api` or `apps/<slug>/go-api`:

- `cmd/api/` — application entry (`main.go`, `app.go` bootstrap).
- `internals/transport/http/handlers/<domain>/` — HTTP handlers, `RegisterEndpoints`, Swagger comments.
- `internals/transport/http/middlewares/` — auth, rate limit, etc.
- `internals/transport/http/gateway.go` — wires handlers to Echo groups.
- `internals/services/<domain>/` — business logic, DTOs, validation, transactions.
- `internals/repositories/<domain>/main.go` — persistence, Squirrel, soft-delete patterns.
- `internals/models/` — shared domain models.
- `internals/constants/` — permissions, feature flags, error codes.
- `internals/jobs/`, `internals/cron/` — async and scheduled work (queue names aligned with root `Makefile` RabbitMQ topology).
- `migrations/` — versioned SQL only.
- `docs/` — `swagger.json`, `swagger.yaml`, generated Go embed.
- `Makefile` — `make test`, `make test-integration`, `make swagger`, `make migrate-create`, etc.

## Canonical example files

- **Gateway registration:** `apps/template/go-api/internals/transport/http/gateway.go`
- **Service (validation, authorizer, errors):** `apps/template/go-api/internals/services/api_key/create.go`
- **Tests:** `apps/template/go-api/internals/services/api_key/main_test.go`
- **Integration-style repo tests:** under `internals/repositories/**/main_test.go`

## go-common usage

- Import paths like `github.com/yca-software/go-common/error`, `…/logger`, `…/validator`, `…/repository`.
- App `go.mod` contains `replace github.com/yca-software/go-common => ../../../packages/go-common`.

## Patterns to follow

- Handlers: bind input, call service, map to HTTP; **Swagger** annotations on exported handler methods.
- Services: `validator.ValidateStruct`, `accessInfo` / authorizer checks, `yca_error` constructors, structured logging.
- Repositories: extend interface in same package; use shared base repo helpers; respect `deleted_at`.
- New domain: add repo → `internals/repositories/repositories.go`, service deps → `internals/services/services.go`, handler → `gateway.go`.

## Anti-patterns

- Business rules or authorization skipped in services “because the handler checks”.
- Ad-hoc schema changes without migrations.
- Forgetting `make swagger` after route/DTO changes.

## Validation

```bash
cd apps/template/go-api   # or apps/<slug>/go-api
make test
# If repositories/migrations touched:
make test-integration
# If HTTP surface changed:
make swagger
```

## Common AI mistakes

- New routes not registered in `gateway.go`.
- Wrong import module prefix after scaffold (`go-api-template` vs `<slug>-go-api`).
- Using `go test` only for packages that need Postgres integration coverage.
