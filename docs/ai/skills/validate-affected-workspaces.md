# Skill: Validate affected workspaces

## When to use

Before finishing any task: run the **minimum** correct checks for everything you touched.

## Pre-checks

- List touched paths (git status or explicit file list).
- Map each path to a workspace using `pnpm-workspace.yaml` (root package, `apps/*`, `apps/*/*`, `packages/*`, `tools/*`).

## Inspection steps

1. For each touched `package.json`, read its `scripts` block — those are authoritative for that package.
2. If only `tools/create-project` changed, see `update-cli-generator.md`.
3. If `go.work` or `packages/go-common` changed, plan Go tests for **all** APIs in `go.work`.

## Implementation steps (commands)

Run from repo root unless `cd` is shown.

### Design system (`packages/design-system`)

```bash
cd packages/design-system && pnpm lint && pnpm test:run
```

### go-common (`packages/go-common`)

```bash
cd packages/go-common && go test ./... -count=1
```

### Go API (`apps/**/go-api`)

```bash
cd apps/<slug>/go-api
make test
# If internals/repositories or migrations changed:
make test-integration
# If HTTP/schema comments changed:
make swagger
```

### React SPA (`apps/**/react-spa`)

```bash
cd apps/<slug>/react-spa
pnpm lint && pnpm test && pnpm build
```

### Marketing (`apps/**/marketing`)

```bash
cd apps/<slug>/marketing
pnpm lint && pnpm build
```

### Expo (`apps/**/expo-mobile`)

```bash
cd apps/<slug>/expo-mobile
pnpm lint && pnpm typecheck
```

### Monorepo-wide (optional / heavy)

```bash
pnpm lint
pnpm build
```

**Note:** `pnpm check-types` runs turbo `check-types`, but many packages do not define that script — do not treat it as sufficient TypeScript coverage.

## Completion checklist

- [ ] Every touched workspace has at least one **executed** validation command recorded in your reasoning.
- [ ] If you skipped integration tests, you have a concrete reason (e.g. no DB changes) and standard `make test` still ran for Go.

## Warnings (this repo)

- Turbo caches builds — if results look stale, rerun with `turbo run build --force` (only when needed).
- Go integration tests may require Docker / Postgres — align with `Makefile` and developer expectations before claiming pass.
