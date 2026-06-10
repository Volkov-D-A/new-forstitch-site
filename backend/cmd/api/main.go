package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"new-forstitch-site/backend/internal/api"
	"new-forstitch-site/backend/internal/db"
	"new-forstitch-site/backend/internal/repository"
	"new-forstitch-site/backend/internal/services"
)

func main() {
	addr := env("HTTP_ADDR", ":3000")
	databaseURL := env("DATABASE_URL", "postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable")
	adminUsername := env("ADMIN_USERNAME", "admin")
	adminPassword := env("ADMIN_PASSWORD", "dev-admin-password")
	allowedOrigins := envList("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Open(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := db.Migrate(ctx, database); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresRepository(database)
	service := services.New(repo)
	if err := service.EnsureAdminUser(adminUsername, adminPassword); err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(service, allowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend api listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envList(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
