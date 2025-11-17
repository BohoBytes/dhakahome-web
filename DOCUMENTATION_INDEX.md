# DhakaHome Web - Documentation Index

This project includes comprehensive documentation to help you understand the architecture, create new pages, and maintain consistency.

## Documents Overview

### 1. **QUICK_START_GUIDE.md** (Start Here!)
   - **Purpose**: Get up and running in 5 minutes
   - **Contains**:
     - Installation & setup instructions
     - Running the development server
     - Creating your first page
     - Creating reusable components
     - Common patterns and troubleshooting
   - **Best for**: New developers, quick reference during development

### 2. **PROJECT_ARCHITECTURE.md** (Deep Dive)
   - **Purpose**: Comprehensive understanding of the entire system
   - **Contains**:
     - Overall project structure
     - Template/view organization (14 sections)
     - Styling approach (Tailwind CSS, fonts, custom components)
     - Template partials and reuse patterns
     - Routing and page rendering mechanisms
     - API client and data integration
     - Design patterns and best practices
     - Creating new pages and components with detailed examples
     - Styling workflow and customization
     - Common patterns and anti-patterns
     - Deployment notes
   - **Best for**: Understanding the "why" behind decisions, reference when implementing complex features

### 3. **ARCHITECTURE_DIAGRAMS.txt** (Visual Reference)
   - **Purpose**: Visual representation of system flows
   - **Contains**:
     - Request flow diagram
     - Template composition hierarchy
     - Data flow for search results
     - HTMX dynamic search flow
     - File organization and responsibilities
     - Styling decision tree
     - Adding new feature checklist
   - **Best for**: Understanding system flow without reading prose, quick visual reference

### 4. **README.md** (Project Info)
   - **Purpose**: Original project README with quick setup
   - **Contains**:
     - Quick start instructions
     - API configuration
     - Deployment instructions
     - Project notes
   - **Best for**: Initial project setup

---

## Quick Navigation

### I want to...

#### Create a new page
1. Read: **QUICK_START_GUIDE.md** → "Creating a New Page" section
2. Reference: **PROJECT_ARCHITECTURE.md** → Section 9
3. Check example: `internal/views/pages/new-home.html`

#### Create a reusable component
1. Read: **QUICK_START_GUIDE.md** → "Creating a Reusable Component" section
2. Reference: **PROJECT_ARCHITECTURE.md** → Section 4
3. Check example: `internal/views/partials/new-services.html`

#### Work with API data
1. Read: **QUICK_START_GUIDE.md** → "Working with Data from API" section
2. Reference: **PROJECT_ARCHITECTURE.md** → Section 6
3. Check example: `internal/handlers/pages.go` → `SearchPage()`

#### Style a component
1. Read: **QUICK_START_GUIDE.md** → "Styling: Three Approaches" section
2. Visual reference: **ARCHITECTURE_DIAGRAMS.txt** → Section 6 "Styling Decision Tree"
3. Reference: **PROJECT_ARCHITECTURE.md** → Section 8

#### Understand the routing
1. Read: **ARCHITECTURE_DIAGRAMS.txt** → Section 1 "Request Flow Diagram"
2. Reference: **PROJECT_ARCHITECTURE.md** → Section 5

#### Deploy the application
1. Read: **QUICK_START_GUIDE.md** → "Deployment" section
2. Reference: **PROJECT_ARCHITECTURE.md** → Section 12

#### Debug a template error
1. Troubleshooting: **QUICK_START_GUIDE.md** → "Troubleshooting" section
2. Understanding templates: **ARCHITECTURE_DIAGRAMS.txt** → Section 2

---

## Key Concepts at a Glance

### Architecture
- **Backend**: Go 1.22 with Chi router
- **Frontend**: Go html/template engine
- **Styling**: Tailwind CSS 3.4 with PostCSS
- **Interactivity**: HTMX 2.0 for dynamic updates
- **Data**: REST API client with mock fallback

### Template Structure
```
base.html (layout)
  └─ pages/*.html (page content)
       └─ partials/*.html (components)
```

### File Organization
| Layer | Location | Purpose |
|-------|----------|---------|
| Routes | `internal/http/router.go` | Define all endpoints |
| Handlers | `internal/handlers/*.go` | Request handling & data prep |
| Templates | `internal/views/pages/*.html` | Page markup |
| Components | `internal/views/partials/*.html` | Reusable HTML blocks |
| API | `internal/api/client.go` | Backend integration |
| Styling | `web/tailwind.input.css` + `tailwind.config.js` | CSS configuration |

### Three Styling Approaches
1. **Inline utilities** (95% of cases): `class="px-6 py-3 bg-primary text-white"`
2. **Component classes** (patterns used 3+ times): `.btn-primary { @apply ... }`
3. **Arbitrary values** (Figma precision): `class="text-[75px]"`

---

## Development Workflow

### Starting the dev server
```bash
# Terminal 1: CSS watch
npm run css:dev

# Terminal 2: Go server
go run ./cmd/web

# Visit http://localhost:5173
```

### Creating a new page (typical workflow)
1. Create `internal/views/pages/my-page.html`
2. Create partials in `internal/views/partials/` as needed
3. Add handler function in `internal/handlers/pages.go`
4. Update render function to include new partials
5. Add route in `internal/http/router.go`
6. Test at `http://localhost:5173/my-page`

### Important Files to Know
- **Routes**: `internal/http/router.go` - Change if you add new endpoints
- **Handlers**: `internal/handlers/pages.go` - Change to fetch data or prepare context
- **Templates**: `internal/views/pages/*.html` - Main content changes
- **Components**: `internal/views/partials/*.html` - Reusable blocks
- **Styling**: `tailwind.config.js` - Add colors, fonts, sizes
- **CSS Components**: `web/tailwind.input.css` - Add reusable CSS classes

---

## Best Practices Summary

### Templates
- Use template composition (base → page → partials)
- Always use full paths in `define` statements
- Pass simple data (avoid complex structs)
- Use custom functions for formatting (e.g., `formatPrice`)

### Styling
- Prefer inline utilities over custom CSS
- Use arbitrary values `[value]` for Figma precision
- Add arbitrary values to safelist in `tailwind.config.js`
- Keep colors, sizes in config, not hardcoded

### Routing
- Define all routes in `router.go`
- Separate page handlers from partial handlers
- Page handlers include all partials; partial handlers return snippets

### API Integration
- Use API client for all backend calls
- Handle errors gracefully (mock fallback available)
- Pass data via simple maps: `map[string]any{}`

### Code Organization
- One handler per page/endpoint
- Keep templates small (extract to partials if >100 lines)
- Keep business logic in handlers, not templates
- Pre-load partials at module init if used frequently (HTMX)

---

## Common Pitfalls to Avoid

1. **Template not rendering**
   - Forgot to add partial to `ParseFiles()` in handler?
   - Template name in `define` doesn't match include path?

2. **Styles not updating**
   - `npm run css:dev` not running?
   - Arbitrary value not in safelist?

3. **API errors not handled**
   - Error is ignored with `_ = cl.SearchProperties()`
   - Should handle and show user-friendly message

4. **Component not reusable**
   - Hardcoded data in template instead of using context?
   - Should extract to partial and pass data via handler

5. **Circular template includes**
   - Don't include a template that includes you back
   - Use composition, not circular patterns

---

## Useful Commands

```bash
# Development
npm run css:dev          # Watch and rebuild CSS
go run ./cmd/web         # Start Go server
go build -o bin/server ./cmd/web  # Build for deployment

# Tailwind CSS
npm run css:build        # One-time CSS build
npm install              # Install dependencies

# Git
git status               # See uncommitted changes
git diff                 # See exact changes
git log --oneline        # See recent commits
```

---

## Project Files Reference

### Core Application
- `/cmd/web/main.go` - Application entry point
- `/internal/http/router.go` - All routes defined here
- `/internal/handlers/pages.go` - Page rendering logic
- `/internal/handlers/partials.go` - HTMX partial endpoints
- `/internal/api/client.go` - Backend API integration

### Templates
- `/internal/views/layouts/base.html` - Main HTML structure
- `/internal/views/pages/` - Page templates
- `/internal/views/partials/` - Reusable components

### Styling
- `/web/tailwind.input.css` - Tailwind configuration & components
- `/tailwind.config.js` - Theme customization
- `/public/assets/tailwind.css` - Generated CSS (don't edit)

### Configuration
- `/go.mod` - Go dependencies
- `/package.json` - Node dependencies
- `/postcss.config.js` - PostCSS configuration
- `/.env` - Environment variables

---

## Getting Help

### For specific tasks
See "Quick Navigation" section above

### For understanding concepts
1. Start with **ARCHITECTURE_DIAGRAMS.txt** for visuals
2. Read **PROJECT_ARCHITECTURE.md** for details
3. Check examples in the actual codebase

### For troubleshooting
1. Check **QUICK_START_GUIDE.md** "Troubleshooting" section
2. Check error messages carefully - they usually point to the issue
3. Verify file paths match template `define` statements

---

## File Sizes Reference

- `PROJECT_ARCHITECTURE.md` - ~25KB (comprehensive, detailed)
- `ARCHITECTURE_DIAGRAMS.txt` - ~15KB (visual, flowcharts)
- `QUICK_START_GUIDE.md` - ~12KB (practical, examples)
- `DOCUMENTATION_INDEX.md` - ~8KB (this file, navigation)

---

## Document Maintenance

These documents are comprehensive and designed to stay current as the project evolves. When making significant changes:

1. Update relevant section in `PROJECT_ARCHITECTURE.md`
2. Update diagrams in `ARCHITECTURE_DIAGRAMS.txt` if flow changes
3. Update examples in `QUICK_START_GUIDE.md` if best practices change
4. Keep this index current with any new documentation

---

**Last Updated**: November 9, 2025
**Project**: DhakaHome Web
**Status**: Production-ready with comprehensive documentation
