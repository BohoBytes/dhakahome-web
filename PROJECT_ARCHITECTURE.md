# DhakaHome Web - Project Architecture Analysis

## Project Overview
This is a Go-based web application for property management in Dhaka, Bangladesh. It uses:
- **Backend**: Go 1.22 with Chi router
- **Frontend**: Go HTML templates with HTMX for interactivity
- **Styling**: Tailwind CSS with PostCSS
- **API Integration**: REST client for backend API
- **Font**: Poppins (Google/custom fonts via WOFF2)

---

## 1. OVERALL PROJECT STRUCTURE

```
dhakahome-web/
├── cmd/
│   └── web/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── client.go           # REST API client (properties, leads)
│   ├── handlers/
│   │   ├── pages.go            # Page rendering handlers
│   │   └── partials.go         # HTMX partial handlers
│   ├── http/
│   │   └── router.go           # Chi router setup
│   ├── mw/
│   │   └── middleware.go       # Request logging middleware
│   └── views/                  # HTML templates
│       ├── layouts/
│       │   └── base.html       # Main page layout
│       ├── pages/              # Full page templates
│       └── partials/           # Reusable components
├── public/
│   └── assets/                 # Static files (CSS, images, fonts, icons)
├── web/
│   └── tailwind.input.css      # Tailwind base + components
├── tailwind.config.js          # Tailwind configuration
├── postcss.config.js           # PostCSS configuration
├── go.mod / go.sum             # Go dependencies
└── package.json                # Node dependencies (Tailwind, PostCSS)
```

**Key Dependencies:**
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/httplog` - Request logging
- Tailwind CSS 3.4.10, PostCSS 8.5.6

---

## 2. TEMPLATE/VIEW ORGANIZATION

### Directory Structure: `internal/views/`

```
internal/views/
├── layouts/
│   └── base.html               # Main HTML wrapper
├── pages/
│   ├── home.html              # Old home page
│   ├── new-home.html          # New home page (Figma design)
│   ├── search.html            # Search page (HTMX-enabled)
│   ├── new-search-results.html # New search results page
│   └── property.html          # Individual property detail page
└── partials/                  # Reusable components
    ├── header.html
    ├── hero.html
    ├── search-box.html
    ├── services.html
    ├── why-dhakahome.html
    ├── properties-by-area.html
    ├── testimonials.html
    ├── results.html            # Search results grid (HTMX-loaded)
    ├── new-hero.html          # New design hero
    ├── new-search.html        # New design search box
    ├── new-services.html      # New design services
    ├── new-why-best.html      # New design value prop
    ├── new-properties-area.html # New design properties by area
    ├── new-testimonials.html  # New design testimonials
    └── new-footer.html        # New design footer
```

### Template Definition Pattern

All templates use Go's `{{define "template/path.html"}}` syntax. Pages are structured in two layers:

**Base Layout** (`base.html`):
```go
{{define "layouts/base.html"}}
<!DOCTYPE html>
<html>
  <head>...</head>
  <body>
    <main>{{ template "content" . }}</main>
    <footer>...</footer>
  </body>
</html>
{{end}}
```

**Page Template** (e.g., `home.html`):
```go
{{define "content"}}
  {{template "partials/header.html" .}}
  {{template "partials/hero.html" .}}
  {{template "partials/search-box.html" .}}
  {{template "partials/services.html" .}}
  {{template "partials/why-dhakahome.html" .}}
  {{template "partials/properties-by-area.html" .}}
  {{template "partials/testimonials.html" .}}
{{end}}

{{define "pages/home.html"}}{{template "layouts/base.html" .}}{{end}}
```

**Key Insight**: 
- Each page defines a `content` template that includes partials
- Pages then delegate to base layout, which renders that content
- This pattern avoids naming collisions and keeps templates modular

---

## 3. STYLING APPROACH

### Framework: Tailwind CSS

**Configuration** (`tailwind.config.js`):
- **Content**: Scans `internal/views/**/*.html` for class detection
- **Theme Customization**:
  - Primary color: `#F44335` (red)
  - Secondary: `#FFE9E8` (light red)
  - Background: `#FFFFFF` (white)
  - Text primary: `#303030` (dark gray)
  - Subtext: `#767676` (medium gray)

**Custom Font Sizes**:
```javascript
fontSize: {
  hero: ["48px", { lineHeight: "1.1" }],
  section: ["36px", { lineHeight: "1.15" }],
  sectionHeader: ["28px", { lineHeight: "1.2" }],
  buttonlg: ["20px", { lineHeight: "1.2" }],
  buttonmd: ["18px", { lineHeight: "1.2" }],
  descTitle: ["20px", { lineHeight: "1.35" }],
  subtext: ["16px", { lineHeight: "1.35" }],
  filter: ["14px", { lineHeight: "1.35" }],
  nav: ["16px", { lineHeight: "1.25" }],
  small: ["12px", { lineHeight: "1.3" }],
}
```

**Custom Font**: Poppins (weights: 400, 500, 600, 700)
- Loaded via WOFF2 files in `public/assets/fonts/`

**Component Layer** (`web/tailwind.input.css`):
```css
@layer components {
  .btn-primary { @apply btn rounded-full bg-primary text-white px-6 py-3 shadow-sm hover:opacity-95; }
  .btn-outline { @apply btn rounded-full border border-textprimary/25 px-6 py-3; }
  .nav-pill { @apply hidden md:flex items-center gap-2 rounded-full bg-white/90 shadow-sm; }
  .search-card { @apply bg-white rounded-[18px] border border-black/10 shadow-[0_6px_24px_rgba(0,0,0,0.08)]; }
  .card-hover { @apply transition hover:shadow-[0_8px_28px_rgba(0,0,0,0.08)]; }
}
```

**Build Process**:
```bash
npm run css:dev    # Watch mode: web/tailwind.input.css → public/assets/tailwind.css
npm run css:build  # Production: same process, one-time
```

**Safelist**: Hardcoded custom classes that Tailwind's content scanner can't detect:
- Dynamic colors, sizes, spacings used in search results
- Examples: `bg-[#f2f2f2]`, `text-[48px]`, `gap-[10px]`

---

## 4. TEMPLATE PARTIALS & COMPONENT REUSE

### Partial Inclusion Pattern

All partials are included via Go's template syntax in page content:
```go
{{template "partials/header.html" .}}
{{template "partials/new-services.html" .}}
```

### Data Flow to Partials

Partials use `.` (the dot operator) to access passed data:
- Most partials receive `nil` (no data needed for static content)
- `results.html` partial receives `{{.List}}` - property list from API
- `new-search-results.html` page receives `{{.List}}` and `{{.Query}}`

### Example: Dynamic Results Partial

**Handler passes data**:
```go
render(w, "pages/new-search-results.html", "new-search-results.html", 
  map[string]any{"List": list, "Query": q})
```

**Template iterates**:
```html
{{range .List.Items}}
  <div class="property-card">
    <h2>{{.Title}}</h2>
    <img src="{{index .Images 0}}" />
    <div>{{.Currency}}{{formatPrice .Price}}</div>
  </div>
{{else}}
  <p>No properties found</p>
{{end}}
```

**Custom Function**:
```go
"formatPrice": formatPrice,  // Adds comma formatting: 105000 → 105,000
```

### Reusable Service Components

**Service Cards** (`new-services.html`):
- Static 4-column grid
- Each card: icon + title + description
- Hardcoded, not data-driven (no partial for reuse)

**Area Cards** (`new-properties-area.html`):
- Static 4-area layout (Gulshan 2, Bashundhara, Malibagh, Uttara)
- Overlaid text on images
- Could be data-driven but currently hardcoded

---

## 5. ROUTING & PAGE RENDERING

### Router Setup: `internal/http/router.go`

Uses **Chi** (lightweight HTTP router):
```go
r := chi.NewMux()

// Static assets
r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("public/assets"))))

// Pages
r.Get("/", handlers.Home)
r.Get("/new-home", handlers.NewHomePage)
r.Get("/search", handlers.SearchPage)
r.Get("/new-search", handlers.NewSearchResultsPage)
r.Get("/properties/{id}", handlers.PropertyPage)

// HTMX Partials
r.Get("/search-partial", handlers.SearchPartial)

// Forms
r.Post("/lead", handlers.SubmitLead)

// Health / Debug
r.Get("/healthz", ...)
r.Get("/debug/api", ...)
```

### Handler Rendering Pattern: `internal/handlers/pages.go`

**Standard Render Function**:
```go
func render(w http.ResponseWriter, topLevelTemplate string, pageFile string, data any) {
  t := template.Must(template.ParseFiles(
    "internal/views/layouts/base.html",
    "internal/views/pages/" + pageFile,
    // Include all required partials
    "internal/views/partials/header.html",
    "internal/views/partials/hero.html",
    "internal/views/partials/search-box.html",
    // ... more partials
  ))
  if err := t.ExecuteTemplate(w, topLevelTemplate, data); err != nil {
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
  }
}
```

**Why Parse All Partials?**
- Go's template system requires all dependencies to be parsed together
- Avoids collisions between similarly-named templates
- Each page pre-loads only the partials it needs

**Page Handlers**:
```go
// Home page
func Home(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  render(w, "pages/home.html", "home.html", nil)
}

// Search page with API data
func SearchPage(w http.ResponseWriter, r *http.Request) {
  q := r.URL.Query()
  cl := api.New()
  list, _ := cl.SearchProperties(q)  // TODO: error handling
  render(w, "pages/search.html", "search.html", 
    map[string]any{"List": list, "Query": q})
}

// Property detail page
func PropertyPage(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  cl := api.New()
  p, _ := cl.GetProperty(id)
  render(w, "pages/property.html", "property.html", 
    map[string]any{"P": p})
}
```

**New Home Page** (separate render function):
```go
func renderNewHome(w http.ResponseWriter, topLevelTemplate string, pageFile string, data any) {
  t := template.Must(template.ParseFiles(
    "internal/views/layouts/base.html",
    "internal/views/pages/" + pageFile,
    "internal/views/partials/new-hero.html",
    "internal/views/partials/new-search.html",
    "internal/views/partials/new-services.html",
    // ... new partials only
  ))
  if err := t.ExecuteTemplate(w, topLevelTemplate, data); err != nil {
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
  }
}

func NewHomePage(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  renderNewHome(w, "pages/new-home.html", "new-home.html", nil)
}
```

**Search Results Handler** (with custom function map):
```go
func NewSearchResultsPage(w http.ResponseWriter, r *http.Request) {
  q := r.URL.Query()
  cl := api.New()
  list, _ := cl.SearchProperties(q)
  
  t := template.Must(template.New("new-search-results.html").Funcs(
    template.FuncMap{"formatPrice": formatPrice},
  ).ParseFiles(
    "internal/views/layouts/base.html",
    "internal/views/pages/new-search-results.html",
  ))
  
  if err := t.ExecuteTemplate(w, "pages/new-search-results.html", 
    map[string]any{"List": list, "Query": q}); err != nil {
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
  }
}
```

### HTMX Partial Rendering: `internal/handlers/partials.go`

Pre-loads partial template once at module init:
```go
var partialT = template.Must(template.ParseFiles("internal/views/partials/results.html"))

func SearchPartial(w http.ResponseWriter, r *http.Request) {
  cl := api.New()
  list, _ := cl.SearchProperties(r.URL.Query())
  _ = partialT.ExecuteTemplate(w, "partials/results.html", 
    map[string]any{"List": list})
}
```

**Used by**: `/search` page with HTMX to update results dynamically without page reload

---

## 6. API CLIENT & DATA INTEGRATION

### API Client: `internal/api/client.go`

**Models**:
```go
type Property struct {
  ID        string
  Title     string
  Address   string
  Price     float64
  Currency  string  // "৳" for Bangladesh Taka
  Images    []string
  Badges    []string  // e.g., ["For Sale", "Verified"]
  Bedrooms  int
  Bathrooms int
  Area      int      // square feet
  Parking   int
}

type PropertyList struct {
  Items []Property
  Page  int
  Pages int
  Total int
}

type LeadReq struct {
  Name         string
  Email        string
  Phone        string
  PropertyID   string
  UTMSource    string
  UTMCampaign  string
  CaptchaToken string
}
```

**Key Methods**:
```go
func (c *Client) SearchProperties(q url.Values) (PropertyList, error)
  // Query params: location, type, area, maxPrice
  // Falls back to mock data if API unavailable

func (c *Client) GetProperty(id string) (Property, error)
  // Fetch single property details

func (c *Client) SubmitLead(in LeadReq) error
  // POST /leads endpoint (implementation pending)
```

**Mock Data Fallback**:
- If API is unreachable or returns 0 results, serves hardcoded mock properties
- 5 mock properties (mix of residential, commercial, office, hostels)
- Filters by location, type, area from query params

**Configuration**:
- Base URL: `API_BASE_URL` env var (default: `http://localhost:3000/api/v1`)
- HTTP timeout: 10 seconds
- Graceful degradation: no errors propagated to user

---

## 7. DESIGN PATTERNS & BEST PRACTICES

### A. Template Composition Pattern

**Problem**: Go templates can't inherit the way web frameworks do. All dependencies must be parsed together.

**Solution**: Use `define` blocks + nested includes
```
base.html (wraps entire page)
  → pages/xxx.html (defines "content")
      → partials/header.html
      → partials/hero.html
      → partials/services.html
```

**Benefits**:
- No naming collisions (each template path is unique)
- Clear separation: layout, page, component
- Easy to add/remove partials per page

### B. Data-First Approach

**Pages that need API data** pass a data map:
```go
map[string]any{"List": list, "Query": q}
```

**Static pages** pass `nil`:
```go
render(w, "pages/home.html", "home.html", nil)
```

**Templates check for data availability**:
```html
{{range .List.Items}}...{{else}}No results{{end}}
```

### C. Custom Template Functions

**Price formatting** (in `NewSearchResultsPage`):
```go
template.FuncMap{"formatPrice": formatPrice}

func formatPrice(price float64) string {
  // 105000 → "105,000"
}
```

Applied in template:
```html
{{.Currency}}{{formatPrice .Price}}
```

### D. HTMX for Dynamic Updates

**On search page**:
```html
<form hx-get="/search-partial" hx-target="#results" hx-push-url="true">
  <input name="q" placeholder="Area or keyword" />
  <button>Apply</button>
</form>
<div id="results">{{ template "partials/results.html" . }}</div>
```

**Flow**:
1. User submits form
2. HTMX intercepts → `GET /search-partial?q=...&type=...`
3. Handler returns just the `<results.html>` partial (no full page)
4. HTMX replaces `#results` div with response
5. URL bar updates (hx-push-url="true")

**Benefits**: Instant feedback without full page reload

### E. Staging Design Migrations

**Old Design** (`home.html`):
- Uses original partials (hero, search-box, services, testimonials)
- Likely being phased out

**New Design** (`new-home.html`):
- Uses `new-*` prefixed partials
- Figma-based design (mentioned in comments)
- Separate render function to load only these partials

**Both coexist**:
- `/` → old home
- `/new-home` → new home
- Allows A/B testing or gradual migration

---

## 8. STYLING BEST PRACTICES OBSERVED

### A. Utility-First CSS (Tailwind)

**Most styles are inline utilities**:
```html
<section class="py-16 bg-white">
  <div class="max-w-7xl mx-auto px-4">
    <h2 class="text-5xl lg:text-[75px] font-medium leading-tight">
```

**Advantages**:
- No CSS file maintenance
- Responsive by default (md:, lg: prefixes)
- Consistent spacing/colors via design tokens

### B. Arbitrary Values for Figma Precision

When Tailwind's defaults don't match design:
```html
<h1 class="text-[75px] font-medium">  <!-- Custom size from Figma -->
<div class="rounded-[20px]">            <!-- Custom radius -->
<img class="w-[337px] h-[300px]">       <!-- Custom dimensions -->
<div class="shadow-[0px_5px_9.9px_0px_rgba(0,0,0,0.15)]">  <!-- Exact shadow -->
```

**Safelist in config** to prevent unused class pruning:
```javascript
safelist: [
  'bg-[#f2f2f2]',
  'text-[48px]',
  'rounded-[20px]',
  // ... all arbitrary values used
]
```

### C. Component Layer for Reusable Classes

`web/tailwind.input.css`:
```css
@layer components {
  .btn-primary { @apply btn rounded-full bg-primary text-white px-6 py-3; }
  .search-card { @apply bg-white rounded-[18px] border border-black/10; }
}
```

**When to use**:
- Repeating patterns (buttons, cards)
- Complex utilities (multiple utilities combined)
- NOT for one-off styles (use inline utilities)

### D. Responsive Design

**Mobile-first approach**:
```html
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
  <!-- 1 column mobile, 2 md, 4 large -->
</div>

<h1 class="text-3xl md:text-5xl lg:text-[75px]">
  <!-- Scale up with viewport -->
</h1>
```

**Key breakpoints**:
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px

### E. Dark Mode Considerations

Not used in current design, but Tailwind supports:
```html
<div class="dark:bg-slate-900">  <!-- Applies if dark mode enabled -->
```

---

## 9. CREATING NEW PAGES & COMPONENTS

### Quick Start: New Page

**Step 1**: Create page template at `internal/views/pages/my-page.html`
```go
{{define "content"}}
<section class="py-16 bg-white">
  {{template "partials/my-component.html" .}}
</section>
{{end}}

{{define "pages/my-page.html"}}{{template "layouts/base.html" .}}{{end}}
```

**Step 2**: Create partials as needed at `internal/views/partials/my-component.html`
```go
{{define "partials/my-component.html"}}
<div class="max-w-7xl mx-auto px-4">
  <h2 class="text-3xl font-bold">My Component</h2>
</div>
{{end}}
```

**Step 3**: Add handler in `internal/handlers/pages.go`
```go
func MyPage(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  render(w, "pages/my-page.html", "my-page.html", nil)
  // OR with data:
  // render(w, "pages/my-page.html", "my-page.html", 
  //   map[string]any{"Data": myData})
}
```

**Step 4**: Add route in `internal/http/router.go`
```go
r.Get("/my-page", handlers.MyPage)
```

**Step 5**: Update render function in `pages.go` if using new partials
```go
func render(w http.ResponseWriter, topLevelTemplate string, pageFile string, data any) {
  t := template.Must(template.ParseFiles(
    "internal/views/layouts/base.html",
    "internal/views/pages/" + pageFile,
    "internal/views/partials/header.html",
    // ... existing partials
    "internal/views/partials/my-component.html",  // Add this
  ))
  // ...
}
```

### Quick Start: New Partial

**Standalone Service Card Component**:
```go
{{define "partials/service-card.html"}}
<div class="bg-white rounded-[10px] border-3 border-gray-700 p-8 text-center">
  <img src="{{.Icon}}" class="w-[120px] h-[120px] mx-auto mb-6" />
  <h3 class="text-2xl font-medium mb-4">{{.Title}}</h3>
  <p class="text-base text-gray-500">{{.Description}}</p>
</div>
{{end}}
```

**Usage in page**:
```go
{{template "partials/service-card.html" (dict "Icon" "/assets/icon1.svg" "Title" "Service 1" "Description" "...")}}
{{template "partials/service-card.html" (dict "Icon" "/assets/icon2.svg" "Title" "Service 2" "Description" "...")}}
```

**Alternative: Data-Driven (with loop)**:

Handler:
```go
services := []map[string]string{
  {"Icon": "/assets/icon1.svg", "Title": "Service 1"},
  {"Icon": "/assets/icon2.svg", "Title": "Service 2"},
}
render(w, "pages/services.html", "services.html", 
  map[string]any{"Services": services})
```

Template:
```go
{{range .Services}}
  {{template "partials/service-card.html" .}}
{{end}}
```

---

## 10. STYLING WORKFLOW

### Development

```bash
# Terminal 1: Watch CSS
npm run css:dev
# Watches web/tailwind.input.css for changes
# Outputs to public/assets/tailwind.css

# Terminal 2: Run Go server
go run ./cmd/web
# Listens on :5173
```

### Adding Custom Styles

**Option A: Inline Utilities (Preferred)**
```html
<button class="px-6 py-3 bg-primary text-white rounded-lg hover:opacity-95">
```

**Option B: Component Layer** (if used 3+ times)
```css
/* web/tailwind.input.css */
@layer components {
  .btn-primary { @apply px-6 py-3 bg-primary text-white rounded-lg hover:opacity-95; }
}
```

```html
<button class="btn-primary">
```

**Option C: Arbitrary Values** (for precise Figma specs)
```html
<h1 class="text-[75px] font-medium leading-tight">
  <!-- Matches exact Figma size -->
</h1>
```

### Adding Custom Colors

In `tailwind.config.js`:
```javascript
theme: {
  extend: {
    colors: {
      "primary": "#F44335",
      "secondary": "#FFE9E8",
      "success": "#4CAF50",  // Add new color
    }
  }
}
```

Then use:
```html
<div class="bg-success text-white">Success message</div>
```

### Adding Custom Fonts

**In web/tailwind.input.css**:
```css
@font-face {
  font-family: "MyFont";
  src: url("/assets/fonts/MyFont.woff2") format("woff2");
  font-weight: 400;
  font-display: swap;
}
```

**In tailwind.config.js**:
```javascript
theme: {
  extend: {
    fontFamily: {
      myfont: ["MyFont", "sans-serif"],
    }
  }
}
```

**In HTML**:
```html
<p class="font-myfont">Uses custom font</p>
```

---

## 11. COMMON PATTERNS & ANTI-PATTERNS

### Good Patterns

✓ **Template composition**: Base → Page → Partials
✓ **Data maps**: `map[string]any{"Key": value}` for flexibility
✓ **Utility-first CSS**: Consistent spacing, colors, responsive
✓ **API client abstraction**: Single `Client` struct, mock fallback
✓ **Handler separation**: Pages vs. Partials (HTMX)
✓ **Router consolidation**: All routes in one `router.go`
✓ **Static assets versioned**: Hashed filenames in git (e.g., `1f002be89...png`)

### Patterns to Avoid

✗ **Don't pass complex structs**: Keep data simple (strings, numbers, slices)
✗ **Avoid logic in templates**: Use functions (formatPrice) instead
✗ **Don't write CSS**: Use Tailwind utilities
✗ **Avoid partial collisions**: Always use full path in `define` (e.g., "partials/header.html")
✗ **Don't ignore errors**: Handlers currently use `_ = cl.SearchProperties()` (should handle)

---

## 12. DEPLOYMENT NOTES

From README:
```bash
# Build
npm ci && npm run css:build && go build -o bin/server ./cmd/web

# Run
./bin/server

# Set API endpoint
export API_BASE_URL=https://your-staging-api
```

**Key env vars**:
- `ADDR` - Server port (default `:5173`)
- `API_BASE_URL` - Backend API endpoint (default `http://localhost:3000/api/v1`)

---

## 13. SUMMARY TABLE

| Aspect | Technology | Location | Notes |
|--------|-----------|----------|-------|
| Backend | Go 1.22 | `cmd/web/main.go` | Entry point |
| Router | Chi v5 | `internal/http/router.go` | Lightweight, modern |
| API Client | Go stdlib | `internal/api/client.go` | Fallback to mocks |
| Templates | Go std `html/template` | `internal/views/` | 3-layer: base, page, partial |
| CSS Framework | Tailwind 3.4 | `tailwind.config.js` | Utility-first |
| CSS Build | PostCSS + Tailwind CLI | `package.json` scripts | Watch mode for dev |
| Interactivity | HTMX 2.0 | HTML `hx-*` attributes | Replaces content without reload |
| Fonts | Poppins (WOFF2) | `public/assets/fonts/` | Custom font stack |
| Assets | Static files | `public/assets/` | Images, icons, CSS |
| Styling | Inline utilities + components | HTML & `tailwind.input.css` | Responsive, arbitrary values |

---

## 14. QUICK REFERENCE: ADDING COMPONENTS

### Data-Driven Property Card Component

**Partial** (`internal/views/partials/property-card.html`):
```go
{{define "partials/property-card.html"}}
<div class="bg-white rounded-lg shadow hover:shadow-lg transition">
  <img src="{{index .Images 0}}" class="w-full h-48 object-cover" />
  <div class="p-4">
    <h3 class="text-lg font-bold">{{.Title}}</h3>
    <p class="text-gray-600">{{.Address}}</p>
    <div class="mt-2 flex justify-between items-center">
      <span class="text-2xl font-bold text-primary">{{.Currency}}{{.Price}}</span>
      <a href="/properties/{{.ID}}" class="btn-primary text-sm py-2 px-4">Details</a>
    </div>
  </div>
</div>
{{end}}
```

**Page usage**:
```go
{{range .List.Items}}
  {{template "partials/property-card.html" .}}
{{end}}
```

**Handler** (in `pages.go`):
```go
func PropertiesPage(w http.ResponseWriter, r *http.Request) {
  cl := api.New()
  list, _ := cl.SearchProperties(r.URL.Query())
  render(w, "pages/properties.html", "properties.html", 
    map[string]any{"List": list})
}
```

---

## Conclusion

This is a **well-structured, Go-native web application** using:
- **Modular templates**: Clear separation of concerns
- **Tailwind CSS**: Modern, responsive styling without custom CSS
- **HTMX**: Progressive enhancement for dynamic interactions
- **API-driven**: Backend integration with graceful fallbacks
- **Pragmatic design**: Hardcoded partials for now, ready to become data-driven

The architecture supports:
- Rapid page development (template composition)
- Design system consistency (Tailwind tokens)
- A/B testing (old vs. new home pages)
- Smooth API integration (mock fallbacks)
- Easy deployment (static build, single binary)

