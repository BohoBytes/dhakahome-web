# Multi-Environment Configuration Guide

## Overview

This project supports **4 environments**, each with separate OAuth credentials and API endpoints:

| Environment | API Endpoint | Use Case |
|-------------|--------------|----------|
| 🏠 **Local** | `http://localhost:3000/api/v1` | Local development with Nestlo backend |
| 🧪 **Staging** | `https://staging-api.nestlo.com/api/v1` | Testing new features |
| 🔬 **UAT** | `https://uat-api.nestlo.com/api/v1` | User acceptance testing |
| 🚀 **Production** | `https://api.nestlo.com/api/v1` | Live production environment |

---

## Quick Start

### Using VS Code (Recommended)

1. **Open VS Code** in this project
2. Press `F5` or go to **Run → Start Debugging**
3. **Select an environment** from the dropdown:
   - 🏠 Local Development
   - 🧪 Staging
   - 🔬 UAT
   - 🚀 Production

4. The server starts with the selected environment's configuration!

### Using Command Line

```bash
# Local
go run ./cmd/web  # Uses .env.local by default

# Staging
ENV_FILE=.env.staging go run ./cmd/web

# UAT
ENV_FILE=.env.uat go run ./cmd/web

# Production
ENV_FILE=.env.production go run ./cmd/web
```

---

## Environment Files

Each environment has its own `.env.*` file:

```
.env.local       # Local development (localhost:3000)
.env.staging     # Staging environment
.env.uat         # UAT environment
.env.production  # Production environment
.env.example     # Template (committed to git)
```

---

## Configuration Steps

### 1. Local Environment (Already Configured ✅)

Your `.env.local` is already set up with working credentials:

```bash
API_BASE_URL=http://localhost:3000/api/v1
API_CLIENT_ID=client-fe9fea8a-736b-4f7d-999e-4a619bc200fa
API_CLIENT_SECRET=YhOm52_II6_DQPMtd0lF94JRW1bhoe0g4CzS6ben3Q0
API_AUTH_URL=http://localhost:3000/api/v1/oauth/token
```

### 2. Staging Environment

Edit `.env.staging` and add your staging credentials:

```bash
API_BASE_URL=https://staging-api.nestlo.com/api/v1
API_CLIENT_ID=your-staging-client-id
API_CLIENT_SECRET=your-staging-client-secret
API_AUTH_URL=https://staging-api.nestlo.com/api/v1/oauth/token
```

### 3. UAT Environment

Edit `.env.uat` and add your UAT credentials:

```bash
API_BASE_URL=https://uat-api.nestlo.com/api/v1
API_CLIENT_ID=your-uat-client-id
API_CLIENT_SECRET=your-uat-client-secret
API_AUTH_URL=https://uat-api.nestlo.com/api/v1/oauth/token
```

### 4. Production Environment

Edit `.env.production` and add your production credentials:

```bash
API_BASE_URL=https://api.nestlo.com/api/v1
API_CLIENT_ID=your-production-client-id
API_CLIENT_SECRET=your-production-client-secret
API_AUTH_URL=https://api.nestlo.com/api/v1/oauth/token
```

---

## How It Works

### VS Code Launch Configuration

The `.vscode/launch.json` file contains launch configurations for each environment:

```json
{
  "configurations": [
    {
      "name": "🏠 Local Development",
      "envFile": "${workspaceFolder}/.env.local"
    },
    {
      "name": "🧪 Staging",
      "envFile": "${workspaceFolder}/.env.staging"
    },
    // ... etc
  ]
}
```

When you press F5, VS Code:
1. Loads the selected `.env.*` file
2. Sets environment variables
3. Starts the Go server
4. OAuth automatically connects to the correct backend!

---

## Switching Environments

### In VS Code:

1. Click the dropdown next to the "Run" button (top toolbar)
2. Select the environment you want
3. Press F5 or click "Run"

### From Terminal:

```bash
# Method 1: Symlink (recommended)
ln -sf .env.staging .env
go run ./cmd/web

# Method 2: Direct specification
godotenv -f .env.staging go run ./cmd/web
```

---

## Security Best Practices

### ✅ What's Safe:

- `.env.example` - Template with placeholder values (committed to git)
- `.vscode/launch.json` - Launch configurations (committed to git)

### ❌ Never Commit:

- `.env.local` - Contains real local credentials
- `.env.staging` - Contains staging credentials
- `.env.uat` - Contains UAT credentials
- `.env.production` - Contains production credentials

All `.env.*` files (except `.env.example`) are in `.gitignore`!

---

## Getting OAuth Credentials

For each environment, you need to get OAuth credentials from the Nestlo backend team:

### Request Format:

```
Environment: [Staging/UAT/Production]
Purpose: DhakaHome Web Frontend
Scopes: assets.read
```

### They will provide:

```
CLIENT_ID: client-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
CLIENT_SECRET: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TOKEN_URL: https://[env]-api.nestlo.com/api/v1/oauth/token
```

---

## Testing Each Environment

### Test OAuth Connection:

```bash
# Local
./test-oauth.sh

# Staging (edit test-oauth.sh to use .env.staging)
# UAT (edit test-oauth.sh to use .env.uat)
# Production (edit test-oauth.sh to use .env.production)
```

### Test Frontend:

1. Start server with environment
2. Visit: `http://localhost:5173`
3. Search for properties
4. Check logs for API endpoint being used

---

## Troubleshooting

### Wrong Environment Loading?

**Check which env file is loaded:**
```bash
# In VS Code, check the debug console for:
API Client initialized:
  Base URL: http://localhost:3000/api/v1  # <- Should match your env
```

### OAuth Failing?

**Verify credentials are correct:**
```bash
# Check .env file has real credentials (not placeholders)
cat .env.staging | grep API_CLIENT_ID
# Should show: API_CLIENT_ID=client-xxxxx (real ID, not "your-staging-client-id")
```

### Can't Select Environment in VS Code?

1. Install Go extension: `ms-vscode.go`
2. Reload VS Code
3. Open Run panel (Ctrl+Shift+D / Cmd+Shift+D)
4. Dropdown should show 4 environments

---

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `ENVIRONMENT` | Environment name | `local`, `staging`, `uat`, `production` |
| `ADDR` | Server port | `:5173` |
| `API_BASE_URL` | Nestlo API endpoint | `https://api.nestlo.com/api/v1` |
| `API_CLIENT_ID` | OAuth client ID | `client-xxxxxxxx-xxxx-...` |
| `API_CLIENT_SECRET` | OAuth client secret | `xxxxxxxxxxxxxx...` |
| `API_TOKEN_SCOPE` | OAuth scopes | `assets.read` |
| `API_AUTH_URL` | OAuth token endpoint | `https://.../oauth/token` |
| `API_AUTH_TOKEN` | Static JWT (optional) | Leave empty to use OAuth |

---

## Example Workflow

### Scenario: Testing a new feature

1. **Develop locally** → Press F5, select "🏠 Local Development"
2. **Test on staging** → Press F5, select "🧪 Staging"
3. **UAT testing** → Press F5, select "🔬 UAT"
4. **Deploy to prod** → Press F5, select "🚀 Production"

Each environment has:
- ✅ Separate OAuth credentials
- ✅ Separate API endpoints
- ✅ Separate data/properties
- ✅ Isolated testing

---

## Adding a New Environment

To add a new environment (e.g., `dev`):

1. **Create env file**: `.env.dev`
2. **Add credentials** for dev environment
3. **Update `.vscode/launch.json`**:
   ```json
   {
     "name": "🔧 Dev",
     "envFile": "${workspaceFolder}/.env.dev"
   }
   ```
4. **Add to `.gitignore`**: `.env.dev`
5. **Document** in this file

---

## Summary

✅ **4 environments configured**: Local, Staging, UAT, Production
✅ **Easy switching**: F5 → Select environment → Done
✅ **Secure**: All credentials in `.gitignore`
✅ **Flexible**: Add more environments as needed
✅ **Production-ready**: Each env isolated and secure

**You can now develop locally and deploy to any environment with one click!** 🚀
