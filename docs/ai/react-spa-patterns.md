# React SPA patterns (AI)

## Purpose

Vite + React 19 + TypeScript SPA: TanStack Query, Zustand, react-router, i18next, Biome, Vitest, Playwright, Tailwind v4, workspace design-system.

## Folder structure (canonical)

Under `apps/react-spa`:

- `src/routes/` — `Root.tsx`, `public/`, `private/`, `admin/` route trees and layouts.
- `src/api/` — hooks and fetchers (use existing `useAPIFetchWrapper` patterns).
- `src/components/` — app-specific UI (prefer design-system primitives).
- `src/helpers/` — auth, fetch wrapper, shared utilities.
- `src/states/` — Zustand stores.
- `src/constants/` — env-derived constants, permissions, **`queryKeys.ts`**.
- `src/types/` — DTOs and domain types.
- `src/locales/` — i18n JSON namespaces.
- `vite.config.ts` — dedupe / optimize for `@yca-software/design-system`.
- `tailwind.config.ts` — content paths include `../../../packages/design-system/src/**/*.{js,ts,jsx,tsx}`.

## Canonical example files

- **Settings feature with DS + table patterns:** `apps/react-spa/src/routes/private/Settings/AuditLog/index.tsx`
- **Design-system shell:** `apps/react-spa/src/main.tsx` (`ThemeProvider`, `TooltipProvider`)
- **Package scripts:** `apps/react-spa/package.json` — `lint` (Biome), `test` (Vitest), `build` (`tsc && vite build`)

## Dependency note

- `apps/react-spa/package.json` uses published semver for `"@yca-software/design-system"`.

## Patterns to follow

- Add types → query keys → `src/api` hooks → route UI → locales.
- Reuse private/admin guard patterns; do not bypass permission gates.
- Env: document new `VITE_*` in `.env.example` (required baseline: `VITE_API_URL`, `VITE_APP_NAME` per existing template docs).

## Anti-patterns

- Duplicating design-system components in `src/components` without trying exports from `@yca-software/design-system`.
- Scatter query key strings instead of extending `src/constants/queryKeys.ts`.

## Validation

```bash
cd apps/react-spa
pnpm lint
pnpm test
pnpm build
```

## Common AI mistakes

- Breaking Tailwind content paths when moving `packages/design-system`.
- E2E assumptions without checking `playwright.config` and existing `test:e2e` layout.
- Forgetting i18n when adding user-visible strings.
