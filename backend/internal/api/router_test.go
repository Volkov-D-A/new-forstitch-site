package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/repository"
	"new-forstitch-site/backend/internal/services"
)

func TestProductsEndpoint(t *testing.T) {
	router := testRouter()
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
	router := testRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/products/missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestCreateOrderEndpoint(t *testing.T) {
	router := testRouter()
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
	router := testRouter()
	body := strings.NewReader(`{"items":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"order_empty"`) {
		t.Fatalf("expected structured validation code, got %s", rec.Body.String())
	}
}

func TestAdminEndpointRequiresToken(t *testing.T) {
	router := testRouter()
	body := strings.NewReader(`{"id":"new-category","label":"Новая категория"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/categories", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"session_required"`) {
		t.Fatalf("expected structured auth error code, got %s", rec.Body.String())
	}
}

func TestAdminCreateCategoryEndpoint(t *testing.T) {
	router := testRouter()
	cookie, csrfToken := loginAdmin(t, router)
	body := strings.NewReader(`{"id":"new-category","label":"Новая категория"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/categories", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "new-category") {
		t.Fatalf("expected created category in response, got %s", rec.Body.String())
	}
}

func testRouter() http.Handler {
	service := services.New(repository.NewMemoryRepository())
	if err := service.EnsureAdminUser("admin", "password"); err != nil {
		panic(err)
	}
	return NewRouter(service, []string{"http://localhost:5173"})
}

func loginAdmin(t *testing.T, router http.Handler) (*http.Cookie, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"admin","password":"password"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var login models.LoginResponse
	if err := json.NewDecoder(rec.Body).Decode(&login); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if login.CSRFToken == "" {
		t.Fatalf("expected csrf token in login response")
	}

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == adminSessionCookie {
			return cookie, login.CSRFToken
		}
	}
	t.Fatalf("expected admin session cookie")
	return nil, ""
}
