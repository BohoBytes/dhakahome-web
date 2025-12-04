# DhakaHome Web

A modern property search web application for Dhaka, Bangladesh, built with Go, HTMX, and Tailwind CSS.

## Tech Stack

- **Backend**: Go 1.22+ with Chi router
- **Frontend**: Go HTML templates + HTMX for dynamic interactions
- **Styling**: Tailwind CSS 3.4 with PostCSS
- **API**: OAuth2 client credentials flow with Nestlo backend
- **Font**: Poppins (Google Fonts)

## Configure API
Set `API_BASE_URL` to your Nestlo endpoint (local/staging/prod). Default is `http://localhost:3000/api/v1`.  
Authentication options:
- Provide `API_CLIENT_ID` / `API_CLIENT_SECRET` (and optionally `API_TOKEN_SCOPE`, `API_AUTH_URL`) to fetch OAuth2 client-credential tokens automatically.
- Or set `API_AUTH_TOKEN` to force a static bearer token (bypasses OAuth).

---

## Quick Start (Local Development)

### Prerequisites

- Go 1.22 or higher
- Node.js (for Tailwind CSS compilation)
- Nestlo backend running on `localhost:3000` (or use **Mock Mode** to develop without backend)

### Setup

1. **Clone and install dependencies:**
   ```bash
   npm install
   cp .env.example .env.local
   ```

2. **Configure environment:**

   **Option A: With Backend (Real API)**

   Edit `.env.local` with your OAuth credentials (already configured for local):
   ```bash
   MOCK_ENABLED=false  # Use real API
   API_BASE_URL=http://localhost:3000/api/v1
   API_CLIENT_ID=client-fe9fea8a-736b-4f7d-999e-4a619bc200fa
   API_CLIENT_SECRET=YhOm52_II6_DQPMtd0lF94JRW1bhoe0g4CzS6ben3Q0
   ```

   **Option B: Mock Mode (No Backend Required)** 🎭

   Edit `.env.local` to enable mock mode:
   ```bash
   MOCK_ENABLED=true  # Use mock data - no backend needed!
   ```

   📖 **See [docs/MOCK_MODE.md](docs/MOCK_MODE.md) for complete mock mode documentation**

3. **Start development:**

   **Terminal 1** (Tailwind CSS watch):
   ```bash
   npm run css:dev
   ```

   **Terminal 2** (Go server):
   ```bash
   go run ./cmd/web
   ```

   Or in **VS Code**: Press `F5` → Select "🏠 Local Development"

4. **Visit:** http://localhost:5173

---

## Multi-Environment Support

This project supports **4 environments** with separate configurations:

| Environment | File | API Endpoint | Status |
|-------------|------|--------------|--------|
| 🏠 **Local** | `.env.local` | `localhost:3000` | ✅ Configured |
| 🧪 **Staging** | `.env.staging` | `staging-api.nestlo.com` | ⚙️ Needs credentials |
| 🔬 **UAT** | `.env.uat` | `uat-api.nestlo.com` | ⚙️ Needs credentials |
| 🚀 **Production** | `.env.production` | `api.nestlo.com` | ⚙️ Needs credentials |

### Switching Environments

**In VS Code:**
1. Press `F5` (or Run → Start Debugging)
2. Select environment from dropdown
3. Server starts with that environment's config

**From Terminal:**
```bash
go run ./cmd/web                         # Uses .env.local (default)
ENV_FILE=.env.staging go run ./cmd/web   # Uses .env.staging
ENV_FILE=.env.uat go run ./cmd/web       # Uses .env.uat
ENV_FILE=.env.production go run ./cmd/web # Uses .env.production
```

📖 **See [docs/ENVIRONMENTS.md](docs/ENVIRONMENTS.md) for complete guide**

---

## API Authentication

### OAuth2 Client Credentials (Automatic)

The application uses OAuth2 client credentials flow for secure server-to-server authentication:

✅ **Automatic token generation** - No manual login required
✅ **Auto token refresh** - Refreshes 2 minutes before expiration
✅ **Secure** - Credentials stored server-side only
✅ **Multi-environment** - Separate credentials per environment

**Configuration:**
```bash
# In .env.local (or other env file)
API_CLIENT_ID=your-client-id
API_CLIENT_SECRET=your-client-secret
API_AUTH_URL=http://localhost:3000/api/v1/oauth/token
API_AUTH_TOKEN=  # Leave empty to use OAuth
```

**How it works:**
1. First API request → OAuth token requested automatically
2. Token cached in memory (15 min expiration)
3. Token auto-refreshes at 13 minutes
4. All API calls use Bearer token header

### Getting OAuth Credentials

Request credentials from the Nestlo backend team for each environment:
- **Local**: Already configured ✅
- **Staging**: Request staging credentials
- **UAT**: Request UAT credentials
- **Production**: Request production credentials

---

## Project Structure

```
dhakahome-web/
├── cmd/web/              # Application entry point
│   └── main.go           # Server initialization
├── internal/
│   ├── api/              # API client with OAuth
│   │   └── client.go     # REST client, property search
│   ├── handlers/         # HTTP handlers
│   │   ├── pages.go      # Full page rendering
│   │   └── partials.go   # HTMX partial rendering
│   ├── http/             # Router configuration
│   │   └── router.go     # Chi router setup
│   └── views/            # HTML templates
│       ├── layouts/      # Base layouts
│       ├── pages/        # Full pages
│       └── partials/     # Reusable components
├── public/assets/        # Static files (CSS, images, fonts)
├── web/                  # Tailwind source
│   └── tailwind.input.css
├── docs/                 # Documentation
├── .env.local            # Local environment config
├── .env.staging          # Staging environment config
├── .env.uat              # UAT environment config
└── .env.production       # Production environment config
```

---

## Development

### CSS Development

Tailwind CSS is used for all styling. Custom utilities are defined in `web/tailwind.input.css`.

```bash
# Watch mode (auto-compile on changes)
npm run css:dev

# Production build (minified)
npm run css:build
```

### Template Development

Templates use Go's `html/template` with a 3-layer architecture:

1. **Base Layout** (`layouts/base.html`) - HTML wrapper
2. **Pages** (`pages/*.html`) - Page-specific content
3. **Partials** (`partials/*.html`) - Reusable components

Example:
```go
{{define "content"}}
  {{template "partials/header.html" .}}
  {{template "partials/search-box.html" .}}
{{end}}
```

### Adding New Pages

1. Create template: `internal/views/pages/my-page.html`
2. Add handler in `internal/handlers/pages.go`
3. Add route in `internal/http/router.go`
4. Update render function to include any new partials

---

## Deployment

### Build

```bash
# Install dependencies and build
npm ci
npm run css:build
go build -o bin/server ./cmd/web
```

### Run

```bash
./bin/server
```

### Environment Variables (Production)

```bash
ADDR=:8080
API_BASE_URL=https://api.nestlo.com/api/v1
API_CLIENT_ID=production-client-id
API_CLIENT_SECRET=production-client-secret
API_AUTH_URL=https://api.nestlo.com/api/v1/oauth/token
```

---

## Features

### Property Search
- Search by location, type, area, price range
- Real-time results with HTMX (no page reload)
- Pagination support
- Property details page

### UI Components
- Responsive design (mobile-first)
- Custom property cards
- Search filters
- Hero section with background image
- Services showcase
- Testimonials section

### API Integration
- OAuth2 authentication
- Property search endpoint: `GET /api/v1/assets`
- Query parameters: location, types, status, price_min, price_max, bedrooms, bathrooms, etc.
- Automatic token refresh

---

## Configuration Reference

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `ADDR` | Server address | `:5173` | No (default: `:5173`) |
| `ENVIRONMENT` | Environment name | `local`, `staging`, `uat`, `production` | No |
| `MOCK_ENABLED` | Use mock data instead of API | `true`, `false` | No (default: `false`) |
| `API_BASE_URL` | Nestlo API endpoint | `http://localhost:3000/api/v1` | Yes* |
| `API_CLIENT_ID` | OAuth client ID | `client-xxxxx...` | Yes* |
| `API_CLIENT_SECRET` | OAuth client secret | `xxxxx...` | Yes* |
| `API_TOKEN_SCOPE` | OAuth scopes | `assets.read` | No (default: `assets.read`) |
| `API_AUTH_URL` | OAuth token endpoint | `http://localhost:3000/api/v1/oauth/token` | No (auto-derived) |
| `API_AUTH_TOKEN` | Static JWT token | `eyJhbGc...` | No (leave empty for OAuth) |

*Not required when `MOCK_ENABLED=true`

---

## Documentation

- **[MOCK_MODE.md](docs/MOCK_MODE.md)** - Mock mode guide (develop without backend)
- **[ENVIRONMENTS.md](docs/ENVIRONMENTS.md)** - Multi-environment configuration guide
- **[PROJECT_ARCHITECTURE.md](docs/PROJECT_ARCHITECTURE.md)** - Detailed architecture documentation

---

## Troubleshooting

### OAuth token errors?
- Check credentials in `.env.local` are correct
- Verify backend OAuth endpoint is accessible
- Check logs for: `✅ OAuth token obtained successfully`

### Wrong API endpoint?
- Check server logs for: `✅ Loaded environment: [name] (from .env.[name])`
- Verify correct `.env.*` file is being loaded

### CSS not updating?
- Make sure `npm run css:dev` is running in Terminal 1
- Check `public/assets/tailwind.css` is being regenerated
- Hard refresh browser (Ctrl+Shift+R / Cmd+Shift+R)

### Port already in use?
- Kill existing process: `lsof -ti:5173 | xargs kill -9`
- Or change port in `.env.local`: `ADDR=:5174`

---

## License

Proprietary - BohoBytes
