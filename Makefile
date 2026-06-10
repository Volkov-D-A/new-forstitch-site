COMPOSE := docker compose
DB_SERVICE := postgres
DATABASE_URL ?= postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable
GOCACHE ?= /tmp/go-build-cache
GOMODCACHE ?= /tmp/go-mod-cache
ADMIN_USERNAME ?= admin
ADMIN_PASSWORD ?= dev-admin-password
CORS_ALLOWED_ORIGINS ?= http://localhost:5174,http://127.0.0.1:5174

.PHONY: backend-run db-migrate db-reset db-start db-stop frontend-run

db-start:
	$(COMPOSE) up -d $(DB_SERVICE)

db-stop:
	$(COMPOSE) stop $(DB_SERVICE)

db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d $(DB_SERVICE)

db-migrate:
	cd backend && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate

backend-run:
	cd backend && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" DATABASE_URL="$(DATABASE_URL)" ADMIN_USERNAME="$(ADMIN_USERNAME)" ADMIN_PASSWORD="$(ADMIN_PASSWORD)" CORS_ALLOWED_ORIGINS="$(CORS_ALLOWED_ORIGINS)" go run ./cmd/api

frontend-run:
	cd frontend && npm run dev -- --host 0.0.0.0
