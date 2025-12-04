package handlers

import (
	"net/http"
	"os"
	"path/filepath"
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

func SubmitLead(w http.ResponseWriter, r *http.Request) {
	// TODO: parse form or JSON, call api.New().SubmitLead(...)
	w.WriteHeader(http.StatusNoContent)
}
