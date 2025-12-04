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
		"eq":          func(a, b any) bool { return a == b },
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
	render(w, "pages/home.html", "home.html", map[string]any{
		"List":        api.PropertyList{},
		"ShowResults": false,
		"ActivePage":  "home",
	})
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cl := api.New()
	list, _ := cl.SearchProperties(q) // TODO: handle error, flash message
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/search-results.html").Funcs(template.FuncMap{
		"eq":          func(a, b any) bool { return a == b },
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
	if err := t.ExecuteTemplate(w, "pages/search-results.html", map[string]any{
		"List":        list,
		"Query":       q,
		"ActivePage":  "home",
		"ShowResults": true,
	}); err != nil {
		log.Printf("search page template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func PropertyPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cl := api.New()
	p, _ := cl.GetProperty(id) // TODO: handle error

	similarQuery := url.Values{}
	if p.Type != "" {
		similarQuery.Set("type", p.Type)
	}
	// grab a few extra so filtering by listing type still yields rows
	similarQuery.Set("limit", "12")

	similar := api.PropertyList{}
	if list, err := cl.SearchProperties(similarQuery); err == nil {
		filtered := make([]api.Property, 0, len(list.Items))
		for _, item := range list.Items {
			if item.ID == p.ID {
				continue
			}
			if p.ListingType != "" && !strings.EqualFold(item.ListingType, p.ListingType) {
				continue
			}
			filtered = append(filtered, item)
		}

		if len(filtered) == 0 {
			for _, item := range list.Items {
				if item.ID == p.ID {
					continue
				}
				filtered = append(filtered, item)
			}
		}

		if len(filtered) > 6 {
			filtered = filtered[:6]
		}

		similar = api.PropertyList{
			Items: filtered,
			Page:  1,
			Pages: 1,
			Total: len(filtered),
		}
	}

	render(w, "pages/property.html", "property.html", map[string]any{
		"P":               p,
		"Similar":         similar,
		"SearchBoxLayout": "static",
		"ShowSimilar":     len(similar.Items) > 0,
		"ActivePage":      "home",
		"SimilarType":     p.Type,
		"SimilarListing":  p.ListingType,
	})
}

func FAQPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("FAQ handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/faq.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
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
	t := template.Must(template.New("pages/about-us.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/about-us.html",
		"internal/views/partials/header.html",
	))
	if err := t.ExecuteTemplate(w, "pages/about-us.html", map[string]any{"ActivePage": "about"}); err != nil {
		log.Printf("About Us template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HotelsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hotels page handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/hotels.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/hotels.html",
		"internal/views/partials/header.html",
	))
	if err := t.ExecuteTemplate(w, "pages/hotels.html", map[string]any{"ActivePage": "hotels"}); err != nil {
		log.Printf("Hotels template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func formatPrice(price float64) string {
	if price == 0 {
		return "0"
	}
	s := strconv.FormatInt(int64(price), 10)
	if len(s) <= 3 {
		return s
	}

	parts := []string{s[len(s)-3:]}
	prefix := s[:len(s)-3]

	for len(prefix) > 0 {
		if len(prefix) <= 2 {
			parts = append(parts, prefix)
			break
		}
		parts = append(parts, prefix[len(prefix)-2:])
		prefix = prefix[:len(prefix)-2]
	}

	var out strings.Builder
	for i := len(parts) - 1; i >= 0; i-- {
		out.WriteString(parts[i])
		if i > 0 {
			out.WriteRune(',')
		}
	}
	return out.String()
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
