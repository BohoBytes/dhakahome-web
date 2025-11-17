package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/BohoBytes/dhakahome-web/internal/api"
)

var partialT = template.Must(template.ParseFiles("internal/views/partials/results.html"))

func SearchPartial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	cl := api.New()
	list, err := cl.SearchProperties(r.URL.Query())
	if err != nil {
		log.Printf("search partial error: %v", err)
		http.Error(w, "Search unavailable", http.StatusBadGateway)
		return
	}
	if err := partialT.ExecuteTemplate(w, "partials/results.html", map[string]any{"List": list}); err != nil {
		log.Printf("render partial error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SubmitLead(w http.ResponseWriter, r *http.Request) {
	// TODO: parse form or JSON, call api.New().SubmitLead(...)
	w.WriteHeader(http.StatusNoContent)
}
