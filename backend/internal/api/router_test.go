package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"new-forstitch-site/backend/internal/store"
)

func TestProductsEndpoint(t *testing.T) {
	router := NewRouter(store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "lighthouse_aniva") {
		t.Fatalf("expected seeded product in response, got %s", rec.Body.String())
	}
}

func TestMissingProductEndpoint(t *testing.T) {
	router := NewRouter(store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/api/products/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestCreateOrderEndpoint(t *testing.T) {
	router := NewRouter(store.NewMemoryStore())
	body := strings.NewReader(`{"items":[{"productId":"lighthouse_aniva","quantity":1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "order_") {
		t.Fatalf("expected order id in response, got %s", rec.Body.String())
	}
}

func TestCreateOrderValidation(t *testing.T) {
	router := NewRouter(store.NewMemoryStore())
	body := strings.NewReader(`{"items":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
