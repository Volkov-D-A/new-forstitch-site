COMPOSE := docker compose
DB_SERVICE := postgres
STORAGE_SERVICE := minio
STORAGE_VOLUME := new-forstitch-site_minio_data
DATABASE_URL ?= postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable
GOCACHE ?= /tmp/go-build-cache
GOMODCACHE ?= /tmp/go-mod-cache
ADMIN_USERNAME ?= dimas
ADMIN_PASSWORD ?= dimas
CORS_ALLOWED_ORIGINS ?= http://localhost:5173,http://127.0.0.1:5173
MINIO_ENDPOINT ?= localhost:9000
MINIO_ACCESS_KEY ?= forstitch
MINIO_SECRET_KEY ?= forstitch-secret
MINIO_BUCKET ?= forstitch
MINIO_USE_SSL ?= false
FILE_BASE_URL ?= http://localhost:3000/api/files

.PHONY: backend-run db-migrate db-reset db-start db-stop frontend-run storage-reset storage-start storage-stop

db-start:
	$(COMPOSE) up -d $(DB_SERVICE)

db-stop:
	$(COMPOSE) stop $(DB_SERVICE)

db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d $(DB_SERVICE)

storage-start:
	$(COMPOSE) up -d $(STORAGE_SERVICE) minio-init

storage-stop:
	$(COMPOSE) stop $(STORAGE_SERVICE)

storage-reset:
	$(COMPOSE) stop $(STORAGE_SERVICE) minio-init
	docker volume rm $(STORAGE_VOLUME)
	$(COMPOSE) up -d $(STORAGE_SERVICE) minio-init

db-migrate:
	cd backend && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate

backend-run:
	cd backend && GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" DATABASE_URL="$(DATABASE_URL)" ADMIN_USERNAME="$(ADMIN_USERNAME)" ADMIN_PASSWORD="$(ADMIN_PASSWORD)" CORS_ALLOWED_ORIGINS="$(CORS_ALLOWED_ORIGINS)" MINIO_ENDPOINT="$(MINIO_ENDPOINT)" MINIO_ACCESS_KEY="$(MINIO_ACCESS_KEY)" MINIO_SECRET_KEY="$(MINIO_SECRET_KEY)" MINIO_BUCKET="$(MINIO_BUCKET)" MINIO_USE_SSL="$(MINIO_USE_SSL)" FILE_BASE_URL="$(FILE_BASE_URL)" go run ./cmd/api

frontend-run:
	cd frontend && npm run dev -- --host 0.0.0.0
