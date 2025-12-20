# Multi-Environment Configuration Guide

## Overview
DhakaHome Web can run against multiple backends (local/staging/UAT/production) using per-environment env files. The server loads env files in this order:
1) `ENV_FILE` (explicitly set)  
2) `.env.local` (if present)  
3) `.env`

Default port is `:5173` unless overridden by `ADDR`.

## Environment Files
- `.env.example` – template checked into git (no secrets)
- `.env.local` – local development
- `.env.staging` – staging
- `.env.uat` – UAT
- `.env.production` – production

All env files except `.env.example` should stay out of version control.

## Running in an Environment
### VS Code (uses existing launch configs)
1. Open the Run/Debug panel.
2. Pick one of: 🏠 Local Development, 🧪 Staging, 🔬 UAT, 🚀 Production.
3. Press F5. VS Code loads the corresponding `.env.*` file.

### Command Line
```bash
# Local (default if .env.local exists)
go run ./cmd/web

# Specify an env file explicitly
ENV_FILE=.env.staging go run ./cmd/web
ENV_FILE=.env.uat go run ./cmd/web
ENV_FILE=.env.production go run ./cmd/web

# Alternative: symlink
ln -sf .env.staging .env
go run ./cmd/web
```

## Required Variables

| Variable | Purpose |
|----------|---------|
| `ENVIRONMENT` | Optional label (local/staging/uat/production) for logs |
| `ADDR` | Listen address (default `:5173`) |
| `API_BASE_URL` | Nestlo API base URL (e.g., `https://staging-api.nestlo.com/api/v1`) |
| `API_AUTH_TOKEN` | Static bearer token (leave empty when using OAuth) |
| `API_CLIENT_ID` | OAuth client ID |
| `API_CLIENT_SECRET` | OAuth client secret |
| `API_TOKEN_SCOPE` | OAuth scope (default `assets.read`) |
| `API_AUTH_URL` | OAuth token URL (derived from `API_BASE_URL` if omitted) |
| `MOCK_ENABLED` | `true/1/yes` forces mock data |
| `CONTACT_EMAIL`, `CONTACT_PHONE_RENT`, `CONTACT_PHONE_SALES`, `PROPERY_ENQUIRY_EMAIL` | Contact defaults for property pages/leads |
| `GTAG_ID`, `META_PIXEL_ID`, `HCAPTCHA_*`, `TURNSTILE_*` | Optional integrations |

Use placeholders in env files committed to git; never commit real credentials.

## Safety & Tips
- Keep secrets in untracked `.env.*` files; double-check `.gitignore` before adding new env files.
- When switching environments often, set `ENV_FILE` in your shell profile or use the symlink approach above.
- The server logs which env file was loaded on startup (`✅ Loaded environment: ...`); check this first when debugging config issues.
- Pair env changes with updates to `docs/ENVIRONMENTS.md` so teammates know how to run the same stack.
