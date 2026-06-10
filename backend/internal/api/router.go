package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/services"
)

type API struct {
	allowedOrigins map[string]struct{}
	service        *services.Service
}

const adminSessionCookie = "forstitch_admin_session"

func NewRouter(service *services.Service, allowedOrigins []string) http.Handler {
	api := &API{
		allowedOrigins: originSet(allowedOrigins),
		service:        service,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("POST /api/auth/login", api.login)
	mux.HandleFunc("GET /api/auth/session", api.session)
	mux.HandleFunc("POST /api/auth/logout", api.logout)
	mux.HandleFunc("GET /api/categories", api.categories)
	mux.HandleFunc("GET /api/products", api.products)
	mux.HandleFunc("GET /api/products/{productID}", api.product)
	mux.HandleFunc("GET /api/gallery", api.gallery)
	mux.HandleFunc("GET /api/blog", api.blog)
	mux.HandleFunc("GET /api/site-content", api.siteContent)
	mux.HandleFunc("POST /api/orders", api.createOrder)
	mux.Handle("GET /api/admin/categories", api.admin(http.HandlerFunc(api.adminCategories)))
	mux.Handle("POST /api/admin/categories", api.admin(http.HandlerFunc(api.createCategory)))
	mux.Handle("PUT /api/admin/categories/{categoryID}", api.admin(http.HandlerFunc(api.updateCategory)))
	mux.Handle("DELETE /api/admin/categories/{categoryID}", api.admin(http.HandlerFunc(api.deleteCategory)))
	mux.Handle("GET /api/admin/products", api.admin(http.HandlerFunc(api.adminProducts)))
	mux.Handle("POST /api/admin/products", api.admin(http.HandlerFunc(api.createProduct)))
	mux.Handle("PUT /api/admin/products/{productID}", api.admin(http.HandlerFunc(api.updateProduct)))
	mux.Handle("DELETE /api/admin/products/{productID}", api.admin(http.HandlerFunc(api.deleteProduct)))

	return api.cors(mux)
}

func (api *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	response, session, expiresAt, err := api.service.Login(req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	setAdminSessionCookie(w, session.ID, expiresAt)
	writeJSON(w, http.StatusOK, response)
}

func (api *API) session(w http.ResponseWriter, r *http.Request) {
	session, err := api.sessionFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusOK, models.SessionResponse{Authenticated: false})
		return
	}
	writeJSON(w, http.StatusOK, models.SessionResponse{
		Authenticated: true,
		Username:      session.Username,
		CSRFToken:     session.CSRFToken,
	})
}

func (api *API) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(adminSessionCookie)
	if err == nil && cookie.Value != "" {
		session, sessionErr := api.service.Session(cookie.Value)
		if sessionErr == nil {
			if csrfErr := api.service.CheckCSRF(session, r.Header.Get("X-CSRF-Token")); csrfErr != nil {
				writeAppError(w, csrfErr)
				return
			}
		}
		_ = api.service.Logout(cookie.Value)
	}
	clearAdminSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) categories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Categories())
}

func (api *API) products(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Products())
}

func (api *API) product(w http.ResponseWriter, r *http.Request) {
	product, err := api.service.Product(r.PathValue("productID"))
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (api *API) gallery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Gallery())
}

func (api *API) blog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Blog())
}

func (api *API) siteContent(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.SiteContent())
}

func (api *API) createOrder(w http.ResponseWriter, r *http.Request) {
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, models.BadRequest("invalid_json", "invalid JSON body"))
		return
	}
	order, err := api.service.CreateOrder(req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (api *API) adminCategories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Categories())
}

func (api *API) createCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if !decodeJSON(w, r, &category) {
		return
	}
	if err := api.service.CreateCategory(category); err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

func (api *API) updateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if !decodeJSON(w, r, &category) {
		return
	}
	if err := api.service.UpdateCategory(r.PathValue("categoryID"), category); err != nil {
		writeAppError(w, err)
		return
	}
	category.ID = r.PathValue("categoryID")
	writeJSON(w, http.StatusOK, category)
}

func (api *API) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if err := api.service.DeleteCategory(r.PathValue("categoryID")); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) adminProducts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Products())
}

func (api *API) createProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if !decodeJSON(w, r, &product) {
		return
	}
	if err := api.service.CreateProduct(product); err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (api *API) updateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if !decodeJSON(w, r, &product) {
		return
	}
	if err := api.service.UpdateProduct(r.PathValue("productID"), product); err != nil {
		writeAppError(w, err)
		return
	}
	product.ID = r.PathValue("productID")
	writeJSON(w, http.StatusOK, product)
}

func (api *API) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := api.service.DeleteProduct(r.PathValue("productID")); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := api.sessionFromRequest(r)
		if err != nil {
			writeAppError(w, err)
			return
		}

		if r.Method != http.MethodGet {
			if err := api.service.CheckCSRF(session, r.Header.Get("X-CSRF-Token")); err != nil {
				writeAppError(w, err)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) sessionFromRequest(r *http.Request) (models.AdminSession, error) {
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return models.AdminSession{}, models.Unauthorized("session_required", "admin session is required")
	}
	return api.service.Session(cookie.Value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeAppError(w, models.BadRequest("invalid_json", "invalid JSON body"))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAppError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	payload := models.ErrorPayloadFrom(err, "internal_error", "internal server error")

	switch {
	case errors.Is(err, models.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, models.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, models.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, models.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, models.ErrValidation):
		status = http.StatusBadRequest
	}

	writeJSON(w, status, models.ErrorResponse{Error: payload})
}

func setAdminSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    sessionID,
		Path:     "/api",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAdminSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    "",
		Path:     "/api",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (api *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if api.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, GET, OPTIONS, POST, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	_, ok := api.allowedOrigins[origin]
	return ok
}

func originSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if origin == "" {
			continue
		}
		set[origin] = struct{}{}
	}
	return set
}
