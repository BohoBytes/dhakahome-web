package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BohoBytes/dhakahome-web/internal/api"
)

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if filepath.Base(wd) == "web" {
		return filepath.Join(wd, "..", "..")
	}
	return wd
}

var partialT = template.Must(template.ParseFiles(
	filepath.Join(getProjectRoot(), "internal/views/partials/results.html"),
))

func SearchPartial(w http.ResponseWriter, r *http.Request) {
	cl := api.New()
	list, err := cl.SearchProperties(r.URL.Query())
	if err != nil {
		log.Printf("search partial error: %v", err)
		http.Error(w, "Search unavailable", http.StatusBadGateway)
		return
	}

	// Add debug headers (visible in browser Network tab → Response Headers)
	addAPIDebugHeaders(w, cl)

	w.Header().Set("Content-Type", "text/html")
	if err := partialT.ExecuteTemplate(w, "partials/results.html", map[string]any{"List": list}); err != nil {
		log.Printf("render partial error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// addAPIDebugHeaders adds debug information about the last API call to response headers
// You can see these in Browser DevTools → Network tab → Response Headers
func addAPIDebugHeaders(w http.ResponseWriter, cl *api.Client) {
	w.Header().Set("X-API-URL", cl.LastRequestURL)
	w.Header().Set("X-API-Status", fmt.Sprintf("%d", cl.LastResponseStatus))
	w.Header().Set("X-API-Duration", fmt.Sprintf("%dms", cl.LastRequestDuration.Milliseconds()))
	if cl.LastResponseError != nil {
		w.Header().Set("X-API-Error", cl.LastResponseError.Error())
	}
}

func SubmitLead(w http.ResponseWriter, r *http.Request) {
	// TODO: parse form or JSON, call api.New().SubmitLead(...)
	w.WriteHeader(http.StatusNoContent)
}
