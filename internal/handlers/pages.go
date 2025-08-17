package handlers

import (
	"html/template"
	"log"
	"net/http"

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
	render(w, "pages/home.html", "home.html", nil)
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cl := api.New()
	list, _ := cl.SearchProperties(q) // TODO: handle error, flash message
	render(w, "pages/search.html", "search.html", map[string]any{"List": list, "Query": q})
}

func PropertyPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cl := api.New()
	p, _ := cl.GetProperty(id) // TODO: handle error
	render(w, "pages/property.html", "property.html", map[string]any{"P": p})
}
