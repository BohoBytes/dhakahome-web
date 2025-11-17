# DhakaHome Web (Go + HTMX + Tailwind)

**Generated:** 2025-08-15T07:10:41.165952Z

## Quick Start (Local)
1. Install Go (>=1.22) and Node.js.
2. `cp .env.example .env` and adjust if needed.
3. Terminal A: `npm install && npm run css:dev`
4. Terminal B: `go run ./cmd/web`
5. Visit http://localhost:5173

## Configure API
Set `API_BASE_URL` to your Nestlo endpoint (local/staging/prod). Default is `http://localhost:3000/api/v1`.  
Authentication options:
- Provide `API_CLIENT_ID` / `API_CLIENT_SECRET` (and optionally `API_TOKEN_SCOPE`, `API_AUTH_URL`) to fetch OAuth2 client-credential tokens automatically.
- Or set `API_AUTH_TOKEN` to force a static bearer token (bypasses OAuth).

## Deploy (Render/Fly/Cloud Run)
- Build command: `npm ci && npm run css:build && go build -o bin/server ./cmd/web`
- Start command: `./bin/server`
- Env: `API_BASE_URL=https://your-staging-api`

## Notes
- CORS is open in dev. Restrict in production.
- Lead form posts to `/lead` handler which should call Nestlo `POST /leads` once endpoint is ready.
- HTMX powers partial updates for `/search` via `/search-partial` endpoint.
- Templates live under `internal/views`.
