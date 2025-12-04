# DhakaHome Web - Quick Start Guide

## Project at a Glance

- **Language**: Go 1.22 backend + Go templates frontend
- **Styling**: Tailwind CSS 3.4
- **Router**: Chi v5
- **Interactivity**: HTMX 2.0
- **API**: REST client with mock fallback
- **Structure**: MVC-like (handlers → templates → partials)

---

## Installation & Running

### Prerequisites
- Go 1.22+
- Node.js 18+ (for Tailwind CSS)

### Setup
```bash
# 1. Clone and enter project
cd /Users/shahriartanvir/Projects/boho/dhakahome-web

# 2. Install Node dependencies
npm install

# 3. Copy environment variables
cp .env.example .env
# Edit .env if needed (API_BASE_URL, ADDR)
```

### Development (Two Terminals)

**Terminal 1 - CSS Watch**:
```bash
npm run css:dev
# Watches web/tailwind.input.css for changes
# Rebuilds → public/assets/tailwind.css
```

**Terminal 2 - Go Server**:
```bash
go run ./cmd/web
# Server runs on http://localhost:5173
```

Visit **http://localhost:5173** in your browser.

---

## File Organization Quick Reference

| Purpose | Location | Note |
|---------|----------|------|
| Routes | `internal/http/router.go` | All endpoints defined here |
| Page handlers | `internal/handlers/pages.go` | Full page rendering |
| Partial handlers | `internal/handlers/partials.go` | HTMX snippets |
| API client | `internal/api/client.go` | Backend integration |
| Main layout | `internal/views/layouts/base.html` | HTML wrapper |
| Page content | `internal/views/pages/*.html` | Page-specific markup |
| Components | `internal/views/partials/*.html` | Reusable blocks |
| CSS base | `web/tailwind.input.css` | Tailwind setup, components |
| CSS config | `tailwind.config.js` | Colors, fonts, theme |
| Static files | `public/assets/` | Images, icons, fonts |

---

## Creating a New Page

### 1. Create Template (`internal/views/pages/about.html`)
```go
{{define "content"}}
<section class="py-16 bg-white">
  <div class="max-w-7xl mx-auto px-4">
    <h1 class="text-3xl font-bold mb-6">About Us</h1>
    <p class="text-lg text-gray-600">Company description here...</p>
  </div>
</section>
{{end}}

{{define "pages/about.html"}}{{template "layouts/base.html" .}}{{end}}
```

### 2. Add Handler (`internal/handlers/pages.go`)
```go
func AboutPage(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/html")
  render(w, "pages/about.html", "about.html", nil)
}
```

### 3. Add Route (`internal/http/router.go`)
```go
r.Get("/about", handlers.AboutPage)
```

### 4. Test
```bash
# Server should auto-reload (if using hot reload)
# Visit http://localhost:5173/about
```

---

## Creating a Reusable Component

### Example: Team Member Card

**1. Create partial** (`internal/views/partials/team-member.html`):
```go
{{define "partials/team-member.html"}}
<div class="bg-white rounded-lg shadow-md p-6 text-center">
  <img src="{{.Image}}" alt="{{.Name}}" class="w-32 h-32 rounded-full mx-auto mb-4" />
  <h3 class="text-xl font-bold">{{.Name}}</h3>
  <p class="text-gray-600">{{.Role}}</p>
  <p class="text-sm text-gray-500 mt-2">{{.Bio}}</p>
</div>
{{end}}
```

**2. Use in page** (`internal/views/pages/team.html`):
```go
{{define "content"}}
<section class="py-16 bg-white">
  <div class="max-w-7xl mx-auto px-4">
    <h2 class="text-3xl font-bold mb-12 text-center">Our Team</h2>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      {{range .Members}}
        {{template "partials/team-member.html" .}}
      {{end}}
    </div>
  </div>
</section>
{{end}}

{{define "pages/team.html"}}{{template "layouts/base.html" .}}{{end}}
```

**3. Add handler** (`internal/handlers/pages.go`):
```go
func TeamPage(w http.ResponseWriter, r *http.Request) {
  members := []map[string]string{
    {"Name": "John Doe", "Role": "CEO", "Image": "/assets/john.jpg", "Bio": "10+ years experience"},
    {"Name": "Jane Smith", "Role": "CTO", "Image": "/assets/jane.jpg", "Bio": "Tech innovator"},
  }
  render(w, "pages/team.html", "team.html", 
    map[string]any{"Members": members})
}
```

**4. Add route** (`internal/http/router.go`):
```go
r.Get("/team", handlers.TeamPage)
```

---

## Working with Data from API

### Search Results Example

**Handler** (`internal/handlers/pages.go`):
```go
func SearchPage(w http.ResponseWriter, r *http.Request) {
  q := r.URL.Query()  // Get search params from URL
  cl := api.New()
  
  list, err := cl.SearchProperties(q)
  if err != nil {
    // TODO: handle error, show flash message
  }
  
  render(w, "pages/search.html", "search.html", 
    map[string]any{
      "List": list,
      "Query": q,
    })
}
```

**Template** (`internal/views/pages/search.html`):
```go
{{define "content"}}
<section class="py-16">
  <div class="max-w-7xl mx-auto px-4">
    {{if .List.Items}}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {{range .List.Items}}
          <a href="/properties/{{.ID}}" class="bg-white rounded-lg shadow hover:shadow-lg">
            <img src="{{index .Images 0}}" class="w-full h-48 object-cover rounded-t-lg" />
            <div class="p-4">
              <h3 class="font-bold">{{.Title}}</h3>
              <p class="text-gray-600">{{.Address}}</p>
              <p class="text-xl font-bold text-primary mt-2">{{.Currency}}{{.Price}}</p>
            </div>
          </a>
        {{end}}
      </div>
    {{else}}
      <p class="text-center text-gray-500">No properties found</p>
    {{end}}
  </div>
</section>
{{end}}

{{define "pages/search.html"}}{{template "layouts/base.html" .}}{{end}}
```

---

## Styling: Three Approaches

### 1. Inline Utilities (Default - Use This!)
```html
<button class="px-6 py-3 bg-primary text-white rounded-lg hover:opacity-95 transition">
  Click me
</button>
```

**When**: 95% of cases. Quick, specific, inline.

### 2. Component Classes (For Repeated Patterns)
```css
/* web/tailwind.input.css */
@layer components {
  .btn-primary {
    @apply px-6 py-3 bg-primary text-white rounded-lg 
           hover:opacity-95 transition;
  }
}
```

```html
<button class="btn-primary">Click me</button>
```

**When**: Pattern used 3+ times. Reduces repetition.

### 3. Arbitrary Values (For Figma Precision)
```html
<h1 class="text-[75px] font-medium leading-[1.2]">
  Large Heading
</h1>
```

**When**: Exact pixel values from Figma. 

**IMPORTANT**: Add to safelist in `tailwind.config.js`:
```javascript
safelist: [
  'text-[75px]',  // Add any arbitrary values here
  'rounded-[20px]',
  // ...
]
```

---

## Tailwind Configuration

**Colors** (`tailwind.config.js`):
```javascript
colors: {
  "primary": "#F44335",      // Red button
  "secondary": "#FFE9E8",    // Light pink
  "bg": "#FFFFFF",           // White background
  "textprimary": "#303030",  // Dark text
  "subtext": "#767676",      // Gray text
}
```

Use: `class="text-primary bg-secondary"`

**Font Sizes** (`tailwind.config.js`):
```javascript
fontSize: {
  "hero": ["48px", { lineHeight: "1.1" }],
  "section": ["36px", { lineHeight: "1.15" }],
  "buttonlg": ["20px", { lineHeight: "1.2" }],
}
```

Use: `class="text-hero"` or `class="text-[75px]"`

**Adding a Color**:
```javascript
// tailwind.config.js
theme: {
  extend: {
    colors: {
      "success": "#4CAF50",  // New color
    }
  }
}
```

Then: `class="bg-success text-white"`

---

## Common Patterns

### Responsive Grid
```html
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
  <!-- 1 column on mobile, 2 on tablet, 4 on desktop -->
</div>
```

### Responsive Text
```html
<h1 class="text-2xl md:text-4xl lg:text-5xl font-bold">
  Heading that scales
</h1>
```

### Conditional Rendering
```html
{{if .HasResults}}
  <div>Found {{.Count}} results</div>
{{else}}
  <div>No results found</div>
{{end}}
```

### Looping with Fallback
```html
{{range .Items}}
  <div>{{.Title}} - {{.Price}}</div>
{{else}}
  <p>No items to display</p>
{{end}}
```

### Custom Template Function (In Handler)
```go
template.FuncMap{
  "formatPrice": func(price float64) string {
    return fmt.Sprintf("%.2f", price)
  },
}
```

```html
<!-- In template -->
{{formatPrice .Price}}
```

---

## HTMX for Dynamic Content

### Form with Live Updates

**HTML** (`internal/views/pages/search.html`):
```html
<form hx-get="/search-partial" 
      hx-target="#results"
      hx-push-url="true">
  <input name="q" placeholder="Search..." />
  <button type="submit">Search</button>
</form>
<div id="results">
  {{template "partials/results.html" .}}
</div>
```

**Handler** (`internal/handlers/partials.go`):
```go
func SearchPartial(w http.ResponseWriter, r *http.Request) {
  cl := api.New()
  list, _ := cl.SearchProperties(r.URL.Query())
  
  // Return ONLY the partial, not full page
  partialT.ExecuteTemplate(w, "partials/results.html", 
    map[string]any{"List": list})
}
```

**Flow**:
1. User enters search term
2. HTMX intercepts form submission
3. Makes GET request to `/search-partial?q=...`
4. Handler fetches data and returns `<results.html>` partial
5. HTMX replaces `#results` div with response
6. No page reload needed

---

## Deployment

### Build
```bash
npm ci                          # Install exact dependencies
npm run css:build               # Generate Tailwind CSS
go build -o bin/server ./cmd/web  # Build Go binary
```

### Run
```bash
export API_BASE_URL="https://api.example.com/v1"
export ADDR=":8080"
./bin/server
```

### Environment Variables
| Variable | Default | Purpose |
|----------|---------|---------|
| `ADDR` | `:5173` | Port to listen on |
| `API_BASE_URL` | `http://localhost:3000/api/v1` | Backend endpoint |

---

## Troubleshooting

### Styles not updating?
- Ensure `npm run css:dev` is running
- Check that Tailwind is scanning your template files
- Use DevTools to verify CSS is loaded (`<link rel="stylesheet" href="/assets/tailwind.css">`)

### Template not rendering?
- Check partial is added to `ParseFiles()` in handler
- Verify `{{define "partials/name.html"}}` matches the include path
- Use `template.Must()` - it will error immediately if template fails

### API not responding?
- Check `API_BASE_URL` environment variable
- App falls back to mock data automatically if API fails
- Mock data is served from `internal/api/client.go`

### CSS not including custom classes?
- Add arbitrary values to `safelist` in `tailwind.config.js`
- Rebuild CSS: `npm run css:build`
- Example: `safelist: ['text-[75px]', 'rounded-[20px]']`

---

## File Locations Reference

```
dhakahome-web/
├── cmd/web/main.go                           # ← Start here to understand flow
├── internal/
│   ├── http/router.go                        # ← Add routes here
│   ├── handlers/
│   │   ├── pages.go                          # ← Add page handlers
│   │   └── partials.go                       # ← HTMX endpoints
│   ├── api/client.go                         # ← API models & methods
│   └── views/
│       ├── layouts/base.html                 # ← Main HTML structure
│       ├── pages/                            # ← New page templates here
│       └── partials/                         # ← New components here
├── web/tailwind.input.css                    # ← Add @layer components
├── tailwind.config.js                        # ← Customize theme, add colors
└── public/assets/                            # ← Static files (generated CSS, images, fonts)
```

---

## Next Steps

1. Read `PROJECT_ARCHITECTURE.md` for deep dive into patterns
2. Read `ARCHITECTURE_DIAGRAMS.txt` for visual flow diagrams
3. Check existing pages in `internal/views/pages/` for examples
4. Explore `internal/views/partials/` to see component patterns
5. Run the app locally and experiment with changes

Happy building!
