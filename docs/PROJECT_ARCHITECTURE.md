# DhakaHome Web – Project Architecture

## Overview
- Go 1.22 application served with Chi.
- Go `html/template` views with a shared base layout and reusable partials.
- Tailwind CSS 3.4 via PostCSS (`web/tailwind.input.css` → `public/assets/tailwind.css`).
- API client for Nestlo with first-class mock support.
- Default port `:5173`; env file precedence: `ENV_FILE` > `.env.local` > `.env`.

## Routing & Entry Points
- `cmd/web/main.go`: loads env file, builds router, starts HTTP server.
- `internal/http/router.go` routes:
  - `/` → Home (hero + search box; results shown only after a search)
  - `/search` → Search results page (advanced filters)
  - `/properties` → Listing page (pre-sorted results)
  - `/properties/{id}` → Property details
  - `/hotels`, `/faq`, `/about-us`, `/contact-us` (+ aliases `/about`, `/contact`)
  - `/api/search/cities`, `/api/search/neighborhoods` → JSON for dropdowns
  - `/lead` → Lead submission
  - `/assets/*` → Static files, plus `/healthz` and `/debug/api`

## Rendering Pattern
- **Base layout**: `internal/views/layouts/base.html` renders `<main>{{template "content" .}}</main>` and footer; loads `/assets/tailwind.css` and HTMX (available for progressive enhancement).
- **render helper** (`internal/handlers/pages.go`):
  - Parses `base.html`, the page file, and shared partials:
    `page-header.html`, `header.html`, `hero.html`, `search-box.html`, `search-results-list.html`, `property-card.html`, `property-badge.html`, `property-stats.html`, `pagination.html`, `common-sections.html`, `services.html`, `why-dhakahome.html`, `properties-by-area.html`, `testimonials.html`, `faq.html`.
  - Template functions: `eq`, `formatPrice` (Bangla comma grouping), `add`, `sub`, `seq`.
- **Search page renderer**: custom `ParseFiles` includes `search-advanced-box.html` plus the shared partials above.
- New partials must be added to the relevant `ParseFiles` lists and defined with their full path (`{{define "partials/foo.html"}}`).

## Search Experience
- Home (`/`): hero, search box, marketing sections; `ShowResults` is false until a search is made.
- `/search`: runs `api.SearchProperties` with the query, populates dropdowns via `withSearchData`, and renders `search-results.html` (advanced box, results list, featured areas).
- `/properties`: applies defaults (`limit=24`, `sort_by=price`, `order=desc`), sorts results in-handler for consistency with mock data, and uses the shared render helper.
- Dropdown data: `withSearchData` builds the `Search` struct (type, city, area, price min/max, listing type, beds/baths, parking, serviced/shared, area ranges) using `api.GetCities`/`api.GetNeighborhoods`; JSON endpoints expose cities/neighborhoods to the client.
- Top areas: `withTopAreas` fetches `GetTopNeighborhoods`, shuffles, and renders four featured areas with images and prebuilt search URLs.

## Property Details Flow
- Handler fetches:
  - `GetProperty(id)` for the main listing
  - `GetRequiredDocuments(type)` for a document checklist
  - Similar listings: `SearchProperties` filtered by type/listing type, excluding the current ID, capped at six items
- Contact data: derives from env (`PROPERY_ENQUIRY_EMAIL`, `CONTACT_PHONE_*`) or property fields, normalizes Bangladesh phone numbers.
- Template data keys: `P`, `Similar`, `Documents`, `ContactEmail`, `ContactPhone`, `ShowSimilar`, `SimilarType`, `SimilarListing`, `SearchBoxLayout`.

## API Client (`internal/api/client.go`)
- Config via env: `API_BASE_URL`, `API_AUTH_TOKEN` or OAuth (`API_CLIENT_ID`, `API_CLIENT_SECRET`, `API_AUTH_URL`, `API_TOKEN_SCOPE`), `MOCK_ENABLED`.
- Methods:
  - `SearchProperties(url.Values)` → `PropertyList`
  - `GetProperty(id)` → `Property`
  - `GetRequiredDocuments(assetType)` → `[]Document`
  - `GetTopNeighborhoods(limit, city)` → `[]NeighborhoodStat`
  - `GetCities()` / `GetNeighborhoods(city)`
  - `SubmitLead(LeadReq)`
- Behavior:
  - `MOCK_ENABLED=true|1|yes` forces mock responses.
  - On real API errors/non-200 responses, the client falls back to mocks automatically.
  - Applies default status filter `listed_rental,listed_sale` in search params.
  - Tracks last request URL/status/duration in logs for debugging.
- Mock dataset: 25 curated properties (residential, commercial, hostels) with filtering, pagination, and price/bed/bath/area logic identical to the real client.

## Templates & Partials (current)
- Pages: `home.html`, `search-results.html`, `properties.html`, `property.html`, `hotels.html`, `faq.html`, `about-us.html`, `contact-us.html`.
- Partials:
  - Layout/support: `page-header.html`, `header.html`, `hero.html`, `common-sections.html`
  - Search: `search-box.html`, `search-advanced-box.html`, `search-results-list.html`, `pagination.html`
  - Listings: `property-card.html`, `property-badge.html`, `property-stats.html`
  - Marketing: `services.html`, `why-dhakahome.html`, `properties-by-area.html`, `testimonials.html`

## Styling System
- Tailwind scans `internal/views/**/*.html` and `web/**/*.css`; safelist covers arbitrary values used by search/results/FAQ/property pages.
- Theme tokens: `primary #F44335`, `secondary #FFE9E8`, `bg`, `secbg`, `textprimary`, `subtext`; Poppins font declared in `web/tailwind.input.css`.
- Base styles: 14px root font size, component layer for `btn`, `btn-primary`, `btn-outline`, `nav-pill`, `nav-link`, `chip-*`, `search-card`, `card-hover`.
- Build commands: `npm run css:dev` (watch) and `npm run css:build` (one-shot).

## Lead Capture
- Endpoint: `POST /lead`.
- Accepts JSON or form payloads; validates/normalizes BD phone numbers and email; uses `api.SubmitLead`.
- Responses: JSON body on AJAX/HTMX requests or `204 No Content` on success.

## Best Practices
- Keep handler data maps simple; avoid business logic in templates.
- Register every new partial in `ParseFiles` and use full template paths in `define`.
- Reuse `withSearchData` for any page that exposes filters or echoes queries.
- Safelist arbitrary Tailwind classes before shipping; reserve the component layer for patterns used multiple times.
- Surface API errors where possible, even though the client already falls back to mocks.

## Maintenance Checklist
- Added a route? Update `router.go` and the docs.
- Added/renamed partials? Update `ParseFiles` lists and the partial inventory above.
- Changed search params/filters? Update `withSearchData`, API param builders, and this document.
- Touched the build pipeline? Update this doc and the Quick Start guide.
