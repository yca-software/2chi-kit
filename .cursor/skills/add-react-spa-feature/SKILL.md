---
name: add-react-spa-feature
description: >-
  Adds screens, settings sections, API hooks, or UI flows to the Vite React
  SPA under apps/**/react-spa. Use when working on react-spa routes, TanStack
  Query hooks, Zustand, i18n locales, or private/admin features.
---

# Add React SPA feature

## Flow

1. Mirror closest route under `src/routes/` (e.g. Settings subfeatures).
2. Order: types → `src/constants/queryKeys.ts` → `src/api/` → route UI → `src/locales/`.
3. Prefer `@yca-software/design-system`; keep `vite.config.ts` / `tailwind.config.ts` patterns when touching deps or Tailwind content.

## Validate

`pnpm lint && pnpm test && pnpm build` in the SPA package.

Full steps: [docs/ai/skills/add-react-spa-feature.md](../../../docs/ai/skills/add-react-spa-feature.md)
