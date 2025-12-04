package main

import (
	"log"
	"net/http"
	"os"

	httpx "github.com/BohoBytes/dhakahome-web/internal/http"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment file (.env.local, .env.staging, etc.)
	// Priority: ENV_FILE env var > .env.local > .env
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		// Try .env.local first (for VS Code launch configs)
		if _, err := os.Stat(".env.local"); err == nil {
			envFile = ".env.local"
		} else {
			envFile = ".env"
		}
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("warning: could not load %s: %v", envFile, err)
	} else {
		env := get("ENVIRONMENT", "local")
		log.Printf("✅ Loaded environment: %s (from %s)", env, envFile)
	}

	addr := get("ADDR", ":5173")
	r := httpx.NewRouter()

	log.Printf("🚀 dhakahome-web listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
