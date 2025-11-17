package main

import (
	"log"
	"net/http"
	"os"

	httpx "github.com/BohoBytes/dhakahome-web/internal/http"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: could not load .env: %v", err)
	}
	addr := get("ADDR", ":5173")
	r := httpx.NewRouter()

	log.Printf("dhakahome-web listening on %s", addr)
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
