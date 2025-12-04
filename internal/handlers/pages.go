package handlers

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/BohoBytes/dhakahome-web/internal/api"
	"github.com/go-chi/chi/v5"
)

// render parses ONLY the base layout + the requested page (+ partials as needed),
// so each page can define its own "content" without collisions.
func render(w http.ResponseWriter, topLevelTemplate string, pageFile string, data any) {
	log.Printf("Rendering template: %s with page: %s", topLevelTemplate, pageFile)
	t := template.Must(template.New(pageFile).Funcs(template.FuncMap{
		"formatPrice": formatPrice,
		"add":         add,
		"sub":         sub,
		"seq":         seq,
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/"+pageFile,
		"internal/views/partials/header.html",
		"internal/views/partials/hero.html",
		"internal/views/partials/search-box.html",
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/pagination.html",
		"internal/views/partials/common-sections.html",
		"internal/views/partials/services.html",
		"internal/views/partials/why-dhakahome.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
		"internal/views/partials/faq.html",
		"internal/views/partials/results.html", // safe to include for all pages
	))
	log.Printf("Templates parsed successfully")
	if err := t.ExecuteTemplate(w, topLevelTemplate, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("Template executed successfully")
}

func Home(w http.ResponseWriter, r *http.Request) {
	log.Printf("Home handler called")
	cl := api.New()
	list, err := cl.SearchProperties(url.Values{})
	if err != nil {
		log.Printf("home search error: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	render(w, "pages/home.html", "home.html", map[string]any{"List": list})
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cl := api.New()
	list, _ := cl.SearchProperties(q) // TODO: handle error, flash message
	t := template.Must(template.New("pages/search-results.html").Funcs(template.FuncMap{
		"formatPrice": formatPrice,
		"add":         add,
		"sub":         sub,
		"seq":         seq,
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/search-results.html",
		"internal/views/partials/header.html",
		"internal/views/partials/hero.html",
		"internal/views/partials/search-box.html",
		"internal/views/partials/common-sections.html",
		"internal/views/partials/services.html",
		"internal/views/partials/why-dhakahome.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
		"internal/views/partials/faq.html",
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/pagination.html",
	))
	if err := t.ExecuteTemplate(w, "pages/search-results.html", map[string]any{"List": list, "Query": q}); err != nil {
		log.Printf("search page template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func PropertyPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cl := api.New()
	p, _ := cl.GetProperty(id) // TODO: handle error
	render(w, "pages/property.html", "property.html", map[string]any{"P": p})
}

func FAQPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("FAQ handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/faq.html",
		"internal/views/partials/header.html",
	))
	if err := t.ExecuteTemplate(w, "pages/faq.html", nil); err != nil {
		log.Printf("FAQ template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AboutUsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("About Us handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/about-us.html",
		"internal/views/partials/header.html",
	))
	if err := t.ExecuteTemplate(w, "pages/about-us.html", nil); err != nil {
		log.Printf("About Us template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func formatPrice(price float64) string {
	if price == 0 {
		return "0"
	}
	s := strconv.FormatInt(int64(price), 10)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result strings.Builder
	rem := n % 3
	if rem > 0 {
		result.WriteString(s[:rem])
		if n > rem {
			result.WriteRune(',')
		}
	}
	for i := rem; i < n; i += 3 {
		result.WriteString(s[i : i+3])
		if i+3 < n {
			result.WriteRune(',')
		}
	}
	return result.String()
}

func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }

func seq(start, end int) []int {
	if end < start {
		return []int{}
	}
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}
