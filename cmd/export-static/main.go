package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BohoBytes/dhakahome-web/internal/api"
	httpx "github.com/BohoBytes/dhakahome-web/internal/http"
)

// This command pre-renders the Go templates to static HTML so Netlify
// (or any static host) can serve the site without running the Go server.
// It uses mock data to avoid backend dependencies.
func main() {
	os.Setenv("MOCK_ENABLED", "true")
	if os.Getenv("ENVIRONMENT") == "" {
		os.Setenv("ENVIRONMENT", "uat")
	}

	router := httpx.NewRouter()
	client := api.New()

	// Core pages to export
	pages := []string{
		"/",
		"/search?q=uttara",
		"/faq",
		"/about-us",
		"/hotels",
		"/properties",
		"/contact-us",
	}

	// Export a few property detail pages using mock data
	if list, err := client.SearchProperties(url.Values{}); err == nil {
		for i, prop := range list.Items {
			if i >= 5 { // limit number of detail pages
				break
			}
			pages = append(pages, "/properties/"+prop.ID)
		}
	} else {
		log.Printf("warning: could not load mock properties: %v", err)
	}

	for _, p := range pages {
		if err := renderToFile(router, p); err != nil {
			log.Fatalf("export failed for %s: %v", p, err)
		}
	}

	log.Printf("✅ Export completed. Files written under public/")
}

func renderToFile(h http.Handler, path string) error {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		return fmt.Errorf("status %d", rr.Code)
	}

	outPath := outputPath(path)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(outPath, rr.Body.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	log.Printf("wrote %s from %s", outPath, path)
	return nil
}

func outputPath(p string) string {
	if p == "/" {
		return filepath.Join("public", "index.html")
	}
	clean := strings.TrimPrefix(p, "/")
	if idx := strings.Index(clean, "?"); idx >= 0 {
		clean = clean[:idx]
	}
	clean = strings.TrimSuffix(clean, "/")
	return filepath.Join("public", clean, "index.html")
}
