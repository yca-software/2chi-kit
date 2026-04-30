# Marketing Template (Astro)

Astro-based marketing template for landing pages.

## Why this stack

- Static-first output for SEO and fast load times.
- React integration to reuse `@yca-software/design-system` marketing components.
- React components for fast section-level customization.

## Structure

- `src/pages/index.astro`: English home (`/`).
- `src/pages/fr/index.astro`: French home (`/fr/`).
- `src/components/MarketingHome.astro`: shared landing; receives `locale` and loads copy from JSON.
- `src/i18n/`: locale list (`config.ts`), message loader (`loadMessages.ts`), and `locales/*.json` string catalogs.
- `src/components/LanguageBar.astro`: EN / FR switcher (respects `import.meta.env.BASE_URL`).
- `src/components/SeoHead.astro`: SEO defaults (title, description, canonical, OG, optional `hreflang` alternates).
- `src/components/ProjectLaunches.astro`: reusable launch/projects section (labels passed in from messages).

## Multilingual

- Default language is English at the site root. French is served under `/fr/`.
- To add a locale: copy a catalog (e.g. `src/i18n/locales/en.json` → `de.json`), register it in `loadMessages.ts`, extend `locales` and `isLocale()` in `src/i18n/config.ts`, and add `src/pages/<locale>/index.astro` that renders `<MarketingHome locale="<locale>" />`.
- Set `site` in `astro.config.mjs` so canonical URLs and the sitemap match production.

## Environment

Copy `.env.example` to `.env` and update values.

- `PUBLIC_SITE_URL`: canonical base URL.
- `PUBLIC_SITE_NAME`: brand/site name.
- `PUBLIC_CONTACT_EMAIL`: footer or contact blocks.
- `PUBLIC_GA_MEASUREMENT_ID`: optional GA4 id.
