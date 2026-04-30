# Skill: Add Go API module (vertical slice)

## When to use

New domain surface (HTTP + business logic + persistence) in an API under `apps/**/go-api`.

## Pre-checks

- Read `docs/ai/go-api-patterns.md`.
- Pick a **reference domain** in the same codebase (e.g. `api_key`, `organization`, `audit_log`).

## Inspection steps

1. Open reference:
   - Handler package: `apps/template/go-api/internals/transport/http/handlers/<domain>/`
   - Service: `apps/template/go-api/internals/services/<domain>/` (e.g. `create.go`, `main_test.go`)
   - Repository: `apps/template/go-api/internals/repositories/<domain>/main.go`
2. Read `internals/transport/http/gateway.go` for registration order and group structure.
3. Read `internals/repositories/repositories.go` and `internals/services/services.go` for constructor wiring.
4. If DB changes: inspect recent migration in `migrations/` for conventions.

## Implementation steps

1. Add models / constants if needed (`internals/models`, `internals/constants`).
2. Add repository interface + implementation; register in `repositories.go`.
3. Add service methods + tests; register deps in `services.go`.
4. Add handler + `RegisterEndpoints`; wire in `gateway.go`.
5. Add SQL migration(s) under `migrations/`.
6. Run `make swagger` from the API directory; commit updated `docs/swagger.*`.

## Validation steps

```bash
cd apps/<path>/go-api
make test
make test-integration   # if repositories or SQL behavior changed
```

## Completion checklist

- [ ] All new routes registered and documented in Swagger.
- [ ] Authorizer / `accessInfo` rules match similar domains.
- [ ] Errors use `go-common/error` patterns consistently.
- [ ] Tests updated or added next to changed packages.

## Warnings (this repo)

- Queue names and DLX must match infra expectations if the feature uses jobs (`internals/jobs`).
- Use the app’s **actual** module path in imports (`github.com/yca-software/<slug>-go-api` in generated apps, `go-api-template` in template).
