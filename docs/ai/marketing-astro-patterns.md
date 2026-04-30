# Marketing (Astro) patterns (AI)

## Purpose

Static-first marketing site: Astro 5, React islands, Tailwind v4, `@yca-software/design-system` marketing components, optional blog content collections.

## Folder structure (canonical)

Under `apps/template/marketing` or `apps/<slug>/marketing`:

- `src/pages/` — Astro pages (e.g. `index.astro`).
- `src/layouts/` — e.g. `BaseLayout.astro`.
- `src/components/` — Astro/local components (e.g. `SeoHead.astro`, `ProjectLaunches.astro`).
- `src/styles/global.css` — imports `@yca-software/design-system/styles.css`; `@source` may point at `packages/design-system` for Tailwind.
- `src/content/blog/` — markdown posts when blog is used (frontmatter conventions per existing files).

## Canonical example files

- **Home page using DS marketing exports:** `apps/template/marketing/src/pages/index.astro`
- **SEO wrapper:** `apps/template/marketing/src/components/SeoHead.astro`
- **Scripts:** `apps/template/marketing/package.json` — `build`: `astro check && astro build`, `lint`: Biome

## Dependency note

- Template `package.json` uses `"@yca-software/design-system": "^0.1.1"` (npm semver), while `react-spa` in the same monorepo typically uses `link:../../../packages/design-system`. When debugging version skew, check **which app** you are in.

- **Production-style app:** `apps/yca-marketing/package.json` uses a narrower Biome `lint` glob set — follow the **package you are editing**, not an abstract default.

## Patterns to follow

- Compose `Navigation`, `Hero`, `Footer`, `Section`, etc. from `@yca-software/design-system` before writing bespoke marketing primitives.
- Centralize SEO via shared head component and canonical URLs.
- Keep env usage documented in `.env.example` when adding public Astro env.

## Anti-patterns

- Heavy client-only logic on landing paths that should stay static.
- Diverging from established `BaseLayout` + `SeoHead` without a reason.

## Validation

```bash
cd apps/template/marketing   # or apps/<slug>/marketing
pnpm lint
pnpm build
```

## Common AI mistakes

- Assuming the same Biome scope as `apps/yca-marketing` when editing `apps/template/marketing` (scripts differ).
- Breaking `@source` path depth when renaming `apps/` segments (Tailwind v4).
