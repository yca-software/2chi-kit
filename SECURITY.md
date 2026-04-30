# Security and Publication Checklist

Use this checklist before making the repository public.

## 1) Secrets and sensitive data

- Confirm no real secrets are committed:
  - API keys
  - OAuth client secrets
  - JWT/token secrets
  - database passwords
  - private keys/certificates
- Keep only placeholder values in `*.env.example`.
- Never commit `.env` files; this repository ignores them via `.gitignore`.

## 2) Dependency and path hygiene

- Ensure no machine-local paths are tracked (for example absolute `link:` dependencies).
- Prefer published semver dependencies for public templates.
- Regenerate lockfiles after dependency source changes:

  ```bash
  pnpm install
  ```

## 3) Repository scan commands

Run from repository root:

```bash
git ls-files | xargs grep -nE "BEGIN (RSA|EC|OPENSSH|DSA|PGP) PRIVATE KEY|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|sk_live_[0-9A-Za-z]{20,}" || true
git ls-files | grep -E "(\.pem$|\.key$|\.p12$|\.pfx$|\.env$|\.env\.)" || true
```

If any result is a real secret, rotate it immediately and remove it from git history before publishing.

## 4) Validate workspaces

- Root checks:

  ```bash
  pnpm lint
  pnpm build
  ```

- App-level checks:
  - `apps/react-spa`: `pnpm lint && pnpm test && pnpm build`
  - `apps/astro`: `pnpm lint && pnpm build`
  - `apps/go-api`: `make test`
