include .env
export

COMPOSE := docker compose
GOCACHE ?= /tmp/go-build-cache
GOMODCACHE ?= /tmp/go-mod-cache
TEST_DATABASE_NAME ?= forstitch_test
TEST_DATABASE_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(TEST_DATABASE_NAME)?sslmode=disable
BACKEND_IMAGE := $(if $(strip $(DOCKERHUB_USER)),$(DOCKERHUB_USER)/,)forstitch-backend:$(APP_VERSION)
FRONTEND_IMAGE := $(if $(strip $(DOCKERHUB_USER)),$(DOCKERHUB_USER)/,)forstitch-frontend:$(APP_VERSION)
RELEASE_CHECKS := go-test go-integration-test go-vet govulncheck frontend-ci frontend-build frontend-lint frontend-test npm-audit

.PHONY: backend-image-build backend-image-push backend-run frontend-build frontend-ci frontend-image-build frontend-image-push frontend-lint frontend-run frontend-test go-integration-test go-test go-vet govulncheck integration-db npm-audit release-gate services-reset services-start services-stop

define run-check
	@log="$$(mktemp)"; \
	if $(1) >"$$log" 2>&1; then \
		printf "%-18s ok\n" "$(2):"; \
		rm -f "$$log"; \
	else \
		status=$$?; \
		printf "%-18s failed\n\n" "$(2):" >&2; \
		cat "$$log" >&2; \
		rm -f "$$log"; \
		exit $$status; \
	fi
endef

services-start:
	$(COMPOSE) up -d

services-stop:
	$(COMPOSE) stop

services-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d

integration-db:
	$(COMPOSE) up -d --wait postgres
	$(COMPOSE) exec -T postgres dropdb \
		--username "$(POSTGRES_USER)" \
		--if-exists \
		--force \
		"$(TEST_DATABASE_NAME)"
	$(COMPOSE) exec -T postgres createdb \
		--username "$(POSTGRES_USER)" \
		"$(TEST_DATABASE_NAME)"

go-integration-test: integration-db
	cd backend && GOCACHE="$(GOCACHE)" TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -tags=integration ./internal/repository

backend-run: services-start
	cd backend && go run ./cmd/api

backend-image-build:
	docker build --tag "$(BACKEND_IMAGE)" backend

backend-image-push:
	@test -n "$(DOCKERHUB_USER)" || (echo "DOCKERHUB_USER is required" && exit 1)
	@test -n "$(APP_VERSION)" || (echo "APP_VERSION is required" && exit 1)
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag "$(BACKEND_IMAGE)" \
		--push \
		backend

frontend-image-build:
	docker build --tag "$(FRONTEND_IMAGE)" frontend

frontend-image-push:
	@test -n "$(DOCKERHUB_USER)" || (echo "DOCKERHUB_USER is required" && exit 1)
	@test -n "$(APP_VERSION)" || (echo "APP_VERSION is required" && exit 1)
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag "$(FRONTEND_IMAGE)" \
		--push \
		frontend

frontend-run:
	cd frontend && npm run dev -- --host 0.0.0.0

go-test:
	$(call run-check,cd backend && go test ./...,go-test)

go-vet:
	$(call run-check,cd backend && go vet ./...,go-vet)

govulncheck:
	$(call run-check,cd backend && govulncheck ./...,govulncheck)

frontend-ci:
	$(call run-check,cd frontend && npm ci,frontend-ci)

frontend-build:
	$(call run-check,cd frontend && npm run build,frontend-build)

frontend-lint:
	$(call run-check,cd frontend && npm run lint,frontend-lint)

frontend-test:
	$(call run-check,cd frontend && npm test,frontend-test)

npm-audit:
	$(call run-check,cd frontend && npm audit,npm-audit)

release-gate:
	@set -e; \
	for check in $(RELEASE_CHECKS); do \
		$(MAKE) --no-print-directory "$$check"; \
	done
