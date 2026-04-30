# Marketing (Astro) patterns (AI)

## Purpose

Static-first marketing site: Astro 5, React islands, Tailwind v4, `@yca-software/design-system` marketing components, optional blog content collections.

## Folder structure (canonical)

Under `apps/astro`:

- `src/pages/` — Astro pages (e.g. `index.astro`).
- `src/layouts/` — e.g. `BaseLayout.astro`.
- `src/components/` — Astro/local components (e.g. `SeoHead.astro`, `ProjectLaunches.astro`).
- `src/styles/global.css` — imports `@yca-software/design-system/styles.css`; `@source` may point at `packages/design-system` for Tailwind.
- `src/content/blog/` — markdown posts when blog is used (frontmatter conventions per existing files).

## Canonical example files

- **Home page using DS marketing exports:** `apps/astro/src/pages/index.astro`
- **SEO wrapper:** `apps/astro/src/components/SeoHead.astro`
- **Scripts:** `apps/astro/package.json` — `build`: `astro check && astro build`, `lint`: Biome

## Dependency note

- `apps/astro/package.json` and `apps/react-spa/package.json` both use published semver for `"@yca-software/design-system"`.

## Patterns to follow

- Compose `Navigation`, `Hero`, `Footer`, `Section`, etc. from `@yca-software/design-system` before writing bespoke marketing primitives.
- Centralize SEO via shared head component and canonical URLs.
- Keep env usage documented in `.env.example` when adding public Astro env.

## Anti-patterns

- Heavy client-only logic on landing paths that should stay static.
- Diverging from established `BaseLayout` + `SeoHead` without a reason.

## Validation

```bash
cd apps/astro
pnpm lint
pnpm build
```

## Common AI mistakes

- Assuming another repository's Astro/Biome scripts apply here without checking `apps/astro/package.json`.
- Breaking `@source` path depth when renaming `apps/` segments (Tailwind v4).
