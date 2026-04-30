# Monorepo overview (AI)

## Purpose

Single Turborepo + pnpm workspace for app workspaces under `apps/*`, plus shared local infrastructure and AI documentation.

`go-common` and `design-system` are now maintained in external repositories and are no longer part of this monorepo.

## Layout (actual paths)

| Area | Path |
|------|------|
| Workspace definition | `pnpm-workspace.yaml` — `apps/*` |
| Turbo tasks | `turbo.json` — `build`, `lint`, `check-types`, `dev` |
| Root scripts | `package.json` — `pnpm build`, `pnpm dev`, `pnpm lint`, `pnpm check-types` |
| Product workspaces | `apps/<slug>` (each app owns its own stack and scripts) |
| Local infra | repo root `Makefile`, `infra/docker-compose.yml`, RabbitMQ topology files |
| AI docs and playbooks | `docs/ai/*`, `.cursor/rules/*`, `.cursor/skills/*` |

## Canonical examples

- Use existing app folders under `apps/*` as the reference implementation for new work.
- There is no in-repo `apps/template` baseline anymore.

## Validation commands (by touch)

Run from **monorepo root** unless a path is given.

| Touched area | Commands |
|--------------|----------|
| Root / many workspaces | `pnpm lint`, `pnpm build` |
| `check-types` | `pnpm check-types` (only effective where workspace-level `check-types` scripts exist) |
| Single app workspace | `cd apps/<slug> && pnpm lint && pnpm build` (and run app-specific tests if present) |
| Go service workspace | Run service-local `make` / `go test` commands documented by that service |
| Infra files (`infra/*`, root `Makefile`) | Run the relevant `make` target(s) and verify services boot cleanly |

## Patterns to follow

- Before coding, copy folder and naming patterns from the closest existing app under `apps/*`.
- Keep changes scoped to the target app unless the request is explicitly cross-cutting.
- Treat shared repos (`go-common`, `design-system`) as external dependencies; update versions/imports intentionally.

## Anti-patterns

- Re-introducing `apps/template` or `packages/*` assumptions into docs, scripts, or automation.
- Assuming root `pnpm check-types` validates all code when workspace scripts are missing.
- Cross-app refactors when the request only targets one workspace.
- New Go domain logic inside handlers or SQL in services.

## Common AI mistakes here

- Suggesting `pnpm create-project` or template sync workflows that no longer exist in this repo.
- Referring to `packages/design-system` or `packages/go-common` as local paths.
- Documenting validation steps that depend on non-existent `apps/*/*` workspaces.
- Forgetting to run app-local checks after app-specific edits.
