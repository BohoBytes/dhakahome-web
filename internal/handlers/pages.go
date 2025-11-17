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
	t := template.Must(template.ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/"+pageFile,
		"internal/views/partials/header.html",
		"internal/views/partials/hero.html",
		"internal/views/partials/search-box.html",
		"internal/views/partials/services.html",
		"internal/views/partials/why-dhakahome.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
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
	w.Header().Set("Content-Type", "text/html")
	cl := api.New()
	list, err := cl.SearchProperties(url.Values{})
	if err != nil {
		log.Printf("search error: %v", err)
	}
	render(w, "pages/home.html", "home.html", map[string]any{
		"List":             list,
		"UseSearchPartial": true,
	})
}

func PropertyPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cl := api.New()
	p, _ := cl.GetProperty(id) // TODO: handle error
	render(w, "pages/property.html", "property.html", map[string]any{"P": p})
}

func formatPrice(price float64) string {
	priceStr := strconv.FormatFloat(price, 'f', 0, 64)
	// Add commas for thousands (from right to left)
	if len(priceStr) <= 3 {
		return priceStr
	}
	var result strings.Builder
	runes := []rune(priceStr)
	for i := 0; i < len(runes); i++ {
		if i > 0 && (len(runes)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(runes[i])
	}
	return result.String()
}

// Template helper functions for pagination
func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func seq(start, end int) []int {
	if start > end {
		return []int{}
	}
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Search Results handler called")
	q := r.URL.Query()
	cl := api.New()
	list, _ := cl.SearchProperties(q) // TODO: handle error, flash message

	// Render search results page with all partials from homepage (/)
	log.Printf("Rendering search results template")
	t := template.Must(template.New("search-results.html").Funcs(template.FuncMap{
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
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/pagination.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
	))
	log.Printf("Templates parsed successfully")
	if err := t.ExecuteTemplate(w, "pages/search-results.html", map[string]any{"List": list, "Query": q}); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("Template executed successfully")
}
