---
name: validate-affected-workspaces
description: >-
  Runs the correct lint, test, build, and Go make targets for each touched
  workspace in the 2chi-2mono monorepo before finishing work. Use when completing
  any coding task, before saying done, when verifying a change, CI parity, or
  the user asks to run checks.
---

# Validate affected workspaces

## Must do

1. List paths you changed; map each to a package (`packages/*`, `apps/*/*`, `tools/*`).
2. Run that package’s scripts from its `package.json` (authoritative).
3. For Go APIs: `make test`; if repositories/migrations touched: `make test-integration`; if HTTP surface changed: `make swagger`.
4. If `packages/go-common` or `go.work` changed, run `go test ./...` in `go-common` and `make test` in at least one API in `go.work`.

## Quick commands

| Area | From repo (adjust path) |
|------|-------------------------|
| design-system | `cd packages/design-system && pnpm lint && pnpm test:run` |
| go-common | `cd packages/go-common && go test ./... -count=1` |
| go-api | `cd apps/<slug>/go-api && make test` |
| react-spa | `cd apps/<slug>/react-spa && pnpm lint && pnpm test && pnpm build` |
| marketing | `cd apps/<slug>/marketing && pnpm lint && pnpm build` |
| expo-mobile | `cd apps/<slug>/expo-mobile && pnpm lint && pnpm typecheck` |

**Note:** Root `pnpm check-types` may no-op if workspaces lack a `check-types` script — do not rely on it alone.

Full checklist: [docs/ai/skills/validate-affected-workspaces.md](../../../docs/ai/skills/validate-affected-workspaces.md)
