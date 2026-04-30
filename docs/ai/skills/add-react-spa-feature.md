# Skill: Add React SPA feature

## When to use

New screen, settings section, or data flow in `apps/**/react-spa`.

## Pre-checks

- Read `docs/ai/react-spa-patterns.md`.
- Choose reference route with similar UX (e.g. settings list + drawer pattern).

## Inspection steps

1. Browse `src/routes/private/` or `src/routes/admin/` for a sibling feature (e.g. `Settings/AuditLog/index.tsx`).
2. Open `src/constants/queryKeys.ts` for key naming.
3. Open `src/api/` for hook patterns (`useQuery`, `useMutation`, error handling).
4. Check `src/locales/en/` (and other locales if present) for namespace structure.

## Implementation steps

1. Add or extend types in `src/types/`.
2. Add query keys; implement hooks in `src/api/`.
3. Build UI under `src/routes/...`; import from `@yca-software/design-system` where possible.
4. Add i18n strings under `src/locales/`.
5. Add/extend tests (Vitest) colocated or under existing test layout; e2e only if feature needs browser flow coverage.

## Validation steps

```bash
cd apps/<path>/react-spa
pnpm lint
pnpm test
pnpm build
```

## Completion checklist

- [ ] Query invalidation matches mutations (follow neighboring features).
- [ ] Private/admin routes still behind existing guards.
- [ ] New `VITE_*` vars documented in `.env.example`.

## Warnings (this repo)

- `vite.config.ts` treats `@yca-software/design-system` specially — follow existing `optimizeDeps` / `ssr` patterns when touching deps.
- Tailwind content includes `packages/design-system` — if classes seem missing, verify `tailwind.config.ts` paths.
