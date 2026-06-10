package main

import (
	"context"
	"log"
	"os"
	"time"

	"new-forstitch-site/backend/internal/db"
)

func main() {
	databaseURL := env("DATABASE_URL", "postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable")
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

	log.Println("database migrations applied")
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
