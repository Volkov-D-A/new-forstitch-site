package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

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
	router := testRouterWithCustomer(t)
	cookie := loginCustomer(t, router)
	body := strings.NewReader(`{"items":[{"productId":"lighthouse_aniva","quantity":1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", body)
	req.AddCookie(cookie)
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
	router := testRouterWithCustomer(t)
	cookie := loginCustomer(t, router)
	body := strings.NewReader(`{"items":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", body)
	req.AddCookie(cookie)
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
	body := strings.NewReader(`{"label":"Новая категория"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/categories", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var category models.Category
	if err := json.NewDecoder(rec.Body).Decode(&category); err != nil {
		t.Fatalf("decode category response: %v", err)
	}
	if category.ID == "" {
		t.Fatalf("expected generated category id")
	}
	if category.Label != "Новая категория" {
		t.Fatalf("expected created category label, got %s", category.Label)
	}
}

func TestAdminUpdateSiteSettingsEndpoint(t *testing.T) {
	router := testRouter()
	cookie, csrfToken := loginAdmin(t, router)
	body := strings.NewReader(`{"featuredProductId":"dragon_library"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/site-settings", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"featuredProductId":"dragon_library"`) {
		t.Fatalf("expected updated site settings, got %s", rec.Body.String())
	}
}

func TestAdminCreateTestimonialEndpoint(t *testing.T) {
	router := testRouter()
	cookie, csrfToken := loginAdmin(t, router)
	body := strings.NewReader(`{"name":"Анна","role":"Вышивальщица","img":"","text":"Очень понятная схема."}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/testimonials", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":`) {
		t.Fatalf("expected generated testimonial id, got %s", rec.Body.String())
	}
}

func TestAdminCreateBlogPostEndpoint(t *testing.T) {
	router := testRouter()
	cookie, csrfToken := loginAdmin(t, router)
	body := strings.NewReader(`{"title":"Процесс вышивки","date":"2026-06-11","tag":"Блог","img":"","excerpt":"Короткая заметка о процессе.","content":"Полный текст записи о процессе вышивки."}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/blog", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":`) {
		t.Fatalf("expected generated blog post id, got %s", rec.Body.String())
	}
}

func TestAdminCreateGalleryItemEndpoint(t *testing.T) {
	router := testRouter()
	cookie, csrfToken := loginAdmin(t, router)
	body := strings.NewReader(`{"title":"Отшив маяка","by":"Анна","img":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/gallery", body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":`) {
		t.Fatalf("expected generated gallery item id, got %s", rec.Body.String())
	}
}

func testRouter() http.Handler {
	service := services.New(repository.NewMemoryRepository())
	if err := service.EnsureAdminUser("admin", "password"); err != nil {
		panic(err)
	}
	return NewRouter(service, []string{"http://localhost:5173"})
}

func testRouterWithCustomer(t *testing.T) http.Handler {
	t.Helper()

	repo := repository.NewMemoryRepository()
	service := services.New(repo)
	if err := service.EnsureAdminUser("admin", "password"); err != nil {
		panic(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash customer password: %v", err)
	}
	if _, _, err := repo.EnsureCustomer("buyer@example.com", "Анна", string(hash)); err != nil {
		t.Fatalf("seed customer: %v", err)
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

func loginCustomer(t *testing.T, router http.Handler) *http.Cookie {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/customer/login", strings.NewReader(`{"username":"buyer@example.com","password":"secret123"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected customer login status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == customerSessionCookie {
			return cookie
		}
	}
	t.Fatalf("expected customer session cookie")
	return nil
}
