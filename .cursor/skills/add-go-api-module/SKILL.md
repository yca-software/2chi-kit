---
name: add-go-api-module
description: >-
  Adds a vertical slice (handler, service, repository, migrations, swagger) to
  a Go API under apps/**/go-api. Use when adding endpoints, domains, REST
  routes, services, repos, or API features in this monorepo.
---

# Add Go API module

## Flow

1. Copy patterns from an existing domain in the same API (e.g. `internals/services/api_key`, matching handler tree).
2. Register repository → `internals/repositories/repositories.go`, service deps → `internals/services/services.go`, handler → `internals/transport/http/gateway.go`.
3. SQL only in `migrations/`. After route/DTO changes: `make swagger` in the API directory.
4. Errors/logging/validation: match `go-common` and existing services.

## Validate

`make test`; if repo/SQL touched: `make test-integration`.

Full steps: [docs/ai/skills/add-go-api-module.md](../../../docs/ai/skills/add-go-api-module.md)
