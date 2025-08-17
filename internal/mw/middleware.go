package mw

import (
    "net/http"

    "github.com/go-chi/httplog"
)

func RequestLogger() func(next http.Handler) http.Handler {
    l := httplog.NewLogger("dhakahome-web", httplog.Options{Concise: true})
    return httplog.RequestLogger(l)
}
