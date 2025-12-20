# DhakaHome Web – Quick Start Guide

## Stack Snapshot
- Go 1.22 with Chi router
- Go `html/template` views + shared partials
- Tailwind CSS 3.4 via PostCSS (`web/tailwind.input.css` → `public/assets/tailwind.css`)
- API client with mock support (toggle via `MOCK_ENABLED`)

## Local Setup (5 minutes)
1. **Install dependencies**
   ```bash
   npm install          # Tailwind/PostCSS
   go mod download      # Optional: pre-fetch Go modules
   ```
2. **Create an env file**
   ```bash
   cp .env.example .env.local
   # Adjust API_BASE_URL, tokens, and ADDR if needed
   ```
3. **Run in two terminals**
   ```bash
   # Terminal 1: CSS watcher
   npm run css:dev

   # Terminal 2: Go server (uses ENV_FILE > .env.local > .env)
   go run ./cmd/web
   # or: make run
   ```
4. **Open the site** at http://localhost:5173 (default `ADDR`).

## Build for Deployment
```bash
npm run css:build
go build -o bin/server ./cmd/web
# run with desired env vars set (ADDR, API_BASE_URL, etc.)
```

## Where Things Live
- Routes: `internal/http/router.go`
- Page handlers: `internal/handlers/pages.go`
- Lead + search filter helpers: `internal/handlers/partials.go`, `internal/handlers/search_filters.go`
- Templates: `internal/views/layouts/base.html`, `internal/views/pages/*.html`, `internal/views/partials/*.html`
- Styles: `tailwind.config.js`, `web/tailwind.input.css`
- Assets output: `public/assets/`

## Creating a New Page
1. **Add a template** (`internal/views/pages/my-page.html`):
   ```go
   {{define "content"}}
   <section class="py-12">
     <h1 class="text-3xl font-semibold mb-4">My Page</h1>
   </section>
   {{end}}
   {{define "pages/my-page.html"}}{{template "layouts/base.html" .}}{{end}}
   ```
2. **Add a handler** (`internal/handlers/pages.go`):
   ```go
   func MyPage(w http.ResponseWriter, r *http.Request) {
     render(w, "pages/my-page.html", "my-page.html",
       map[string]any{"ActivePage": "my-page"})
   }
   ```
   If you introduce a new partial, add it to the `ParseFiles` list in `render` (and in the search page renderer if it is used there).
3. **Wire the route** (`internal/http/router.go`):
   ```go
   r.Get("/my-page", handlers.MyPage)
   ```
4. **Use shared data helpers**: wrap handler data with `withSearchData` when you need dropdown options or query echoing.

## Working with Data
- Create a client: `cl := api.New()` (uses env vars or mock mode).
- Core methods:
  - `SearchProperties(q url.Values)`
  - `GetProperty(id)`
  - `GetRequiredDocuments(type)`
  - `GetTopNeighborhoods(limit, city)`
  - `SubmitLead(api.LeadReq)`
- Search helpers: `withSearchData` builds `Search` dropdown options using `GetCities` / `GetNeighborhoods` and normalizes incoming query params.
- Mock behavior: `MOCK_ENABLED=true` forces mock data; the client also falls back to mocks when the real API fails.

## Styling Tips
- Tailwind scans `internal/views/**/*.html` and `web/**/*.css`.
- Add reusable classes in `@layer components` inside `web/tailwind.input.css`.
- Use arbitrary values when needed and safelist them in `tailwind.config.js`.
- Base font size is 14px; Poppins is loaded via WOFF2 in `/assets/fonts/`.

## Troubleshooting
- **CSS not updating**: ensure `npm run css:dev` is running and classes are safelisted when arbitrary values are used.
- **Template parse errors**: confirm every partial used is listed in `ParseFiles` and defined with its full path (`{{define "partials/foo.html"}}`).
- **Env not loading**: set `ENV_FILE` or create `.env.local`; see `ENVIRONMENTS.md`.
- **API unreachable**: verify `API_BASE_URL`; toggle `MOCK_ENABLED=true` to develop offline.

Next: read [PROJECT_ARCHITECTURE.md](PROJECT_ARCHITECTURE.md) for the full system picture and [ENVIRONMENTS.md](ENVIRONMENTS.md) to switch between stacks.
