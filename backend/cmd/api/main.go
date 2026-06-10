package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"new-forstitch-site/backend/internal/api"
	"new-forstitch-site/backend/internal/store"
)

func main() {
	addr := env("HTTP_ADDR", ":3000")

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(store.NewMemoryStore()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend api listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
