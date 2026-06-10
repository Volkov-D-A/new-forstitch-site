package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"new-forstitch-site/backend/internal/site"
	"new-forstitch-site/backend/internal/store"
)

type Store interface {
	Categories() []site.Category
	Products() []site.Product
	Product(id string) (site.Product, error)
	Gallery() []site.GalleryItem
	Blog() []site.BlogPost
	SiteContent() site.SiteContent
	CreateOrder(req site.OrderRequest) site.OrderResponse
}

type API struct {
	store Store
}

func NewRouter(dataStore Store) http.Handler {
	api := &API{store: dataStore}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("GET /api/categories", api.categories)
	mux.HandleFunc("GET /api/products", api.products)
	mux.HandleFunc("GET /api/products/{productID}", api.product)
	mux.HandleFunc("GET /api/gallery", api.gallery)
	mux.HandleFunc("GET /api/blog", api.blog)
	mux.HandleFunc("GET /api/site-content", api.siteContent)
	mux.HandleFunc("POST /api/orders", api.createOrder)

	return cors(mux)
}

func (api *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) categories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.store.Categories())
}

func (api *API) products(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.store.Products())
}

func (api *API) product(w http.ResponseWriter, r *http.Request) {
	product, err := api.store.Product(r.PathValue("productID"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load product")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (api *API) gallery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.store.Gallery())
}

func (api *API) blog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.store.Blog())
}

func (api *API) siteContent(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.store.SiteContent())
}

func (api *API) createOrder(w http.ResponseWriter, r *http.Request) {
	var req site.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateOrder(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, api.store.CreateOrder(req))
}

func validateOrder(req site.OrderRequest) error {
	if len(req.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for _, item := range req.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return errors.New("productId is required")
		}
		if item.Quantity < 1 {
			return errors.New("quantity must be greater than zero")
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
