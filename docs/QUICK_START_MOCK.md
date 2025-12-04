# Quick Start with Mock Mode 🎭

**Develop without a backend in 3 steps!**

## 1. Enable Mock Mode

Edit your `.env.local`:
```bash
MOCK_ENABLED=true
```

## 2. Start the Server

```bash
# Terminal 1: Start Tailwind CSS
npm run css:dev

# Terminal 2: Start Go server
go run cmd/web/main.go
```

## 3. Verify Mock Mode

Check the logs for:
```
🎭 API Client: MOCK MODE ENABLED - All API calls will use mock data
```

Visit: http://localhost:5173

---

## What You Get

✅ **23 Properties** across 8 areas
✅ **All search features** work perfectly
✅ **Pagination** supported
✅ **Zero backend** required
✅ **Fast responses** (no network calls)

## Try These URLs

```bash
# All properties
http://localhost:5173/

# Gulshan properties
http://localhost:5173/search?location=Gulshan

# Commercial properties
http://localhost:5173/search?type=Commercial

# Affordable properties
http://localhost:5173/search?price_max=30000

# Luxury properties
http://localhost:5173/search?price_min=80000

# 3+ bedrooms
http://localhost:5173/search?bedrooms=3

# Page 2
http://localhost:5173/search?page=2&limit=5
```

## Switch to Real API

Change `.env.local`:
```bash
MOCK_ENABLED=false
API_BASE_URL=http://localhost:3000/api/v1
API_CLIENT_ID=your-client-id
API_CLIENT_SECRET=your-client-secret
```

---

**Full documentation**: [docs/MOCK_MODE.md](./MOCK_MODE.md)
