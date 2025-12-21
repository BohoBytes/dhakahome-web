package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/BohoBytes/dhakahome-web/internal/api"
	"github.com/go-chi/chi/v5"
)

type shortlistStatusRequest struct {
	AssetIDs    []string `json:"assetIds"`
	AssetIDsAlt []string `json:"asset_ids"`
}

type shortlistAddPayload struct {
	AssetID    string `json:"assetId"`
	AssetIDAlt string `json:"asset_id"`
}

func shortlistToken(r *http.Request) string {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

func parsePositiveInt(val string, def int) int {
	if def <= 0 {
		def = 1
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func isUnauthorized(err error) bool {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr != nil && apiErr.StatusCode == http.StatusUnauthorized {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unauthorized")
}

// ShortlistStatuses handles bulk shortlist checks for the current user.
func ShortlistStatuses(w http.ResponseWriter, r *http.Request) {
	token := shortlistToken(r)
	if token == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	var payload shortlistStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ids := payload.AssetIDs
	if len(ids) == 0 && len(payload.AssetIDsAlt) > 0 {
		ids = payload.AssetIDsAlt
	}

	if len(ids) == 0 {
		http.Error(w, "assetIds is required", http.StatusBadRequest)
		return
	}

	client := api.New()
	statuses := make([]api.ShortlistStatus, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		status, err := client.CheckShortlist(id, token)
		if err != nil {
			if isUnauthorized(err) {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			http.Error(w, "unable to check shortlist right now", http.StatusBadGateway)
			return
		}
		statuses = append(statuses, status)
	}

	writeJSON(w, map[string]any{
		"statuses": statuses,
	})
}

// AddShortlistItem adds a property to the user's shortlist.
func AddShortlistItem(w http.ResponseWriter, r *http.Request) {
	token := shortlistToken(r)
	if token == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	var payload shortlistAddPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	assetID := strings.TrimSpace(payload.AssetID)
	if assetID == "" {
		assetID = strings.TrimSpace(payload.AssetIDAlt)
	}
	if assetID == "" {
		http.Error(w, "assetId is required", http.StatusBadRequest)
		return
	}

	client := api.New()
	status, err := client.AddToShortlist(assetID, token)
	if err != nil {
		if isUnauthorized(err) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		http.Error(w, "unable to add to shortlist", http.StatusBadGateway)
		return
	}

	writeJSON(w, map[string]any{
		"assetId":       status.AssetID,
		"shortlistId":   status.ShortlistID,
		"shortlisted":   status.IsShortlisted,
		"isShortlisted": status.IsShortlisted,
	})
}

// RemoveShortlistItem removes a property from the user's shortlist.
func RemoveShortlistItem(w http.ResponseWriter, r *http.Request) {
	token := shortlistToken(r)
	if token == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	assetID := strings.TrimSpace(chi.URLParam(r, "assetID"))
	if assetID == "" {
		http.Error(w, "assetID is required", http.StatusBadRequest)
		return
	}

	client := api.New()
	status, err := client.RemoveFromShortlist(assetID, token)
	if err != nil {
		if isUnauthorized(err) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		http.Error(w, "unable to remove from shortlist", http.StatusBadGateway)
		return
	}

	writeJSON(w, map[string]any{
		"assetId":       status.AssetID,
		"shortlistId":   status.ShortlistID,
		"shortlisted":   status.IsShortlisted,
		"isShortlisted": status.IsShortlisted,
	})
}

// ShortlistResultsView renders the shortlist results list for the authenticated user.
func ShortlistResultsView(w http.ResponseWriter, r *http.Request) {
	token := shortlistToken(r)
	if token == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 9)

	client := api.New()
	list, err := client.ListShortlisted(token, page, limit)
	if err != nil {
		if isUnauthorized(err) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		http.Error(w, "unable to load shortlist", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	templates := []string{
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/property-badge.html",
		"internal/views/partials/pagination.html",
	}

	funcs := template.FuncMap{
		"eq":          func(a, b any) bool { return a == b },
		"formatPrice": formatPrice,
		"add":         add,
		"sub":         sub,
		"seq":         seq,
		"dict":        dict,
	}

	t := template.Must(template.New("shortlist-partial").Funcs(funcs).ParseFiles(templates...))

	data := map[string]any{
		"ActivePage":       "search",
		"List":             list,
		"Query":            url.Values{},
		"ShowResults":      true,
		"ShortlistEnabled": true,
		"ShortlistMode":    true,
	}

	if err := t.ExecuteTemplate(w, "partials/search-results-list.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
