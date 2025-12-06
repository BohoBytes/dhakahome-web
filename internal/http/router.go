package httpx

import (
	"net/http"
	"os"

	"github.com/BohoBytes/dhakahome-web/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewMux()

	// Temporarily disable middleware for debugging
	// r.Use(mw.RequestLogger())
	// r.Use(cors.Handler(cors.Options{
	//     AllowedOrigins:   []string{"*"}, // dev only; restrict in prod
	//     AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	//     AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	//     AllowCredentials: false,
	//     MaxAge:           300,
	// }))

	// static assets
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("public/assets"))))

	// pages
	r.Get("/", handlers.Home)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h1>Test page works!</h1>"))
	})
	r.Get("/search", handlers.SearchPage)
	r.Get("/faq", handlers.FAQPage)
	r.Get("/about-us", handlers.AboutUsPage)
	r.Get("/about-us/", handlers.AboutUsPage) // allow trailing slash
	r.Get("/about", handlers.AboutUsPage)     // alias
	r.Get("/hotels", handlers.HotelsPage)
	r.Get("/properties/{id}", handlers.PropertyPage)

	// htmx partials
	// forms
	r.Post("/lead", handlers.SubmitLead)

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// debug api
	r.Get("/debug/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(os.Getenv("API_BASE_URL")))
	})

	return r
}
