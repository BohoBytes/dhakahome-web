package handlers

import (
    "html/template"
    "net/http"

    "github.com/BohoBytes/dhakahome-web/internal/api"
)

var partialT = template.Must(template.ParseFiles("internal/views/partials/results.html"))

func SearchPartial(w http.ResponseWriter, r *http.Request) {
    cl := api.New()
    list, _ := cl.SearchProperties(r.URL.Query())
    _ = partialT.ExecuteTemplate(w, "partials/results.html", map[string]any{"List": list})
}

func SubmitLead(w http.ResponseWriter, r *http.Request) {
    // TODO: parse form or JSON, call api.New().SubmitLead(...)
    w.WriteHeader(http.StatusNoContent)
}
