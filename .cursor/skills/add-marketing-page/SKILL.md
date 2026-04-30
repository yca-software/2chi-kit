---
name: add-marketing-page
description: >-
  Adds or edits Astro marketing pages, layouts, SEO, or design-system marketing
  components under apps/**/marketing. Use when working on landing pages,
  marketing site, Astro, SeoHead, BaseLayout, or static site in this repo.
---

# Add marketing page

## Flow

1. Match the package you edit (`apps/template/marketing` vs `apps/yca-marketing` vs `apps/<slug>/marketing`).
2. Follow `src/pages/`, `src/layouts/BaseLayout.astro`, `src/components/SeoHead.astro`; prefer DS imports like `index.astro`.
3. Run **this** package’s `lint` script (Biome scope differs per app).

## Validate

`pnpm lint && pnpm build` (includes `astro check` in template).

Full steps: [docs/ai/skills/add-marketing-page.md](../../../docs/ai/skills/add-marketing-page.md)
