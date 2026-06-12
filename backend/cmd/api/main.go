package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"new-forstitch-site/backend/internal/api"
	"new-forstitch-site/backend/internal/db"
	"new-forstitch-site/backend/internal/mailer"
	"new-forstitch-site/backend/internal/repository"
	"new-forstitch-site/backend/internal/services"
	"new-forstitch-site/backend/internal/storage"
)

func main() {
	addr := env("HTTP_ADDR", ":3000")
	databaseURL := env("DATABASE_URL", "postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable")
	adminUsername := env("ADMIN_USERNAME", "dimas")
	adminPassword := env("ADMIN_PASSWORD", "dimas")
	allowedOrigins := envList("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")
	minioEndpoint := env("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := env("MINIO_ACCESS_KEY", "forstitch")
	minioSecretKey := env("MINIO_SECRET_KEY", "forstitch-secret")
	minioBucket := env("MINIO_BUCKET", "forstitch")
	minioUseSSL := envBool("MINIO_USE_SSL", false)
	fileBaseURL := env("FILE_BASE_URL", "http://localhost:3000/api/files")
	appBaseURL := env("APP_BASE_URL", "http://localhost:3000")
	mailEnabled := envBool("MAIL_ENABLED", false)
	mailHost := env("MAIL_HOST", "localhost")
	mailPort := env("MAIL_PORT", "1025")
	mailUsername := env("MAIL_USERNAME", "")
	mailPassword := env("MAIL_PASSWORD", "")
	mailFrom := env("MAIL_FROM", "no-reply@forstitch.local")
	mailFromName := env("MAIL_FROM_NAME", "Forstitch")
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
	fileStorage, err := storage.NewMinIO(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, minioUseSSL)
	if err != nil {
		log.Fatal(err)
	}
	if err := fileStorage.EnsureBucket(ctx); err != nil {
		log.Fatal(err)
	}
	service.ConfigureFiles(fileStorage, fileBaseURL)
	if mailEnabled {
		service.ConfigureMailer(mailer.SMTP{
			Host:     mailHost,
			Port:     mailPort,
			Username: mailUsername,
			Password: mailPassword,
			From:     mailFrom,
			FromName: mailFromName,
		}, appBaseURL)
	} else {
		service.ConfigureMailer(mailer.Noop{}, appBaseURL)
	}
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

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
