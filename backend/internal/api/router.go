package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/services"
)

type API struct {
	allowedOrigins map[string]struct{}
	service        *services.Service
}

const adminSessionCookie = "forstitch_admin_session"
const customerSessionCookie = "forstitch_customer_session"

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
	mux.HandleFunc("GET /api/files/{objectName...}", api.file)
	mux.HandleFunc("GET /api/gallery", api.gallery)
	mux.HandleFunc("GET /api/blog", api.blog)
	mux.HandleFunc("GET /api/site-content", api.siteContent)
	mux.Handle("POST /api/orders", api.customer(http.HandlerFunc(api.createOrder)))
	mux.HandleFunc("POST /api/customer/register/start", api.customerRegisterStart)
	mux.HandleFunc("POST /api/customer/register/verify", api.customerRegisterVerify)
	mux.HandleFunc("POST /api/customer/password-reset/start", api.customerPasswordResetStart)
	mux.HandleFunc("POST /api/customer/password-reset/verify", api.customerPasswordResetVerify)
	mux.HandleFunc("POST /api/customer/login", api.customerLogin)
	mux.HandleFunc("GET /api/customer/session", api.customerSession)
	mux.HandleFunc("POST /api/customer/logout", api.customerLogout)
	mux.Handle("GET /api/customer/orders", api.customer(http.HandlerFunc(api.customerOrders)))
	mux.Handle("GET /api/customer/orders/{orderID}", api.customer(http.HandlerFunc(api.customerOrder)))
	mux.Handle("GET /api/admin/categories", api.admin(http.HandlerFunc(api.adminCategories)))
	mux.Handle("POST /api/admin/categories", api.admin(http.HandlerFunc(api.createCategory)))
	mux.Handle("PUT /api/admin/categories/{categoryID}", api.admin(http.HandlerFunc(api.updateCategory)))
	mux.Handle("DELETE /api/admin/categories/{categoryID}", api.admin(http.HandlerFunc(api.deleteCategory)))
	mux.Handle("GET /api/admin/products", api.admin(http.HandlerFunc(api.adminProducts)))
	mux.Handle("POST /api/admin/products", api.admin(http.HandlerFunc(api.createProduct)))
	mux.Handle("PUT /api/admin/products/{productID}", api.admin(http.HandlerFunc(api.updateProduct)))
	mux.Handle("POST /api/admin/products/{productID}/image", api.admin(http.HandlerFunc(api.uploadProductImage)))
	mux.Handle("POST /api/admin/products/{productID}/images", api.admin(http.HandlerFunc(api.uploadProductAdditionalImage)))
	mux.Handle("DELETE /api/admin/products/{productID}/images/{imageID}", api.admin(http.HandlerFunc(api.deleteProductImage)))
	mux.Handle("DELETE /api/admin/products/{productID}", api.admin(http.HandlerFunc(api.deleteProduct)))
	mux.Handle("GET /api/admin/blog", api.admin(http.HandlerFunc(api.adminBlog)))
	mux.Handle("POST /api/admin/blog", api.admin(http.HandlerFunc(api.createBlogPost)))
	mux.Handle("PUT /api/admin/blog/{postID}", api.admin(http.HandlerFunc(api.updateBlogPost)))
	mux.Handle("POST /api/admin/blog/{postID}/image", api.admin(http.HandlerFunc(api.uploadBlogPostImage)))
	mux.Handle("DELETE /api/admin/blog/{postID}", api.admin(http.HandlerFunc(api.deleteBlogPost)))
	mux.Handle("GET /api/admin/gallery", api.admin(http.HandlerFunc(api.adminGallery)))
	mux.Handle("POST /api/admin/gallery", api.admin(http.HandlerFunc(api.createGalleryItem)))
	mux.Handle("PUT /api/admin/gallery/{galleryItemID}", api.admin(http.HandlerFunc(api.updateGalleryItem)))
	mux.Handle("POST /api/admin/gallery/{galleryItemID}/image", api.admin(http.HandlerFunc(api.uploadGalleryItemImage)))
	mux.Handle("DELETE /api/admin/gallery/{galleryItemID}", api.admin(http.HandlerFunc(api.deleteGalleryItem)))
	mux.Handle("GET /api/admin/site-settings", api.admin(http.HandlerFunc(api.adminSiteSettings)))
	mux.Handle("PUT /api/admin/site-settings", api.admin(http.HandlerFunc(api.updateSiteSettings)))
	mux.Handle("GET /api/admin/orders", api.admin(http.HandlerFunc(api.adminOrders)))
	mux.Handle("GET /api/admin/testimonials", api.admin(http.HandlerFunc(api.adminTestimonials)))
	mux.Handle("POST /api/admin/testimonials", api.admin(http.HandlerFunc(api.createTestimonial)))
	mux.Handle("PUT /api/admin/testimonials/{testimonialID}", api.admin(http.HandlerFunc(api.updateTestimonial)))
	mux.Handle("POST /api/admin/testimonials/{testimonialID}/image", api.admin(http.HandlerFunc(api.uploadTestimonialImage)))
	mux.Handle("DELETE /api/admin/testimonials/{testimonialID}", api.admin(http.HandlerFunc(api.deleteTestimonial)))

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

func (api *API) file(w http.ResponseWriter, r *http.Request) {
	reader, info, err := api.service.File(r.Context(), r.PathValue("objectName"))
	if err != nil {
		writeAppError(w, err)
		return
	}
	defer reader.Close()

	if info.ContentType != "" {
		w.Header().Set("Content-Type", info.ContentType)
	}
	if info.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = io.Copy(w, reader)
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

func (api *API) adminSiteSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.SiteSettings())
}

func (api *API) adminOrders(w http.ResponseWriter, _ *http.Request) {
	orders, err := api.service.AdminOrders()
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (api *API) updateSiteSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SiteSettings
	if !decodeJSON(w, r, &settings) {
		return
	}
	updated, err := api.service.UpdateSiteSettings(settings)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (api *API) adminTestimonials(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Testimonials())
}

func (api *API) createTestimonial(w http.ResponseWriter, r *http.Request) {
	var testimonial models.Testimonial
	if !decodeJSON(w, r, &testimonial) {
		return
	}
	created, err := api.service.CreateTestimonial(testimonial)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api *API) updateTestimonial(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("testimonialID"))
	if !ok {
		return
	}

	var testimonial models.Testimonial
	if !decodeJSON(w, r, &testimonial) {
		return
	}
	updated, err := api.service.UpdateTestimonial(id, testimonial)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (api *API) uploadTestimonialImage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("testimonialID"))
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeAppError(w, models.BadRequest("invalid_multipart", "invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAppError(w, models.BadRequest("file_required", "file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	testimonial, err := api.service.UploadTestimonialImage(r.Context(), id, header.Filename, contentType, file, header.Size)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, testimonial)
}

func (api *API) deleteTestimonial(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("testimonialID"))
	if !ok {
		return
	}
	if err := api.service.DeleteTestimonial(id); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) createOrder(w http.ResponseWriter, r *http.Request) {
	session, err := api.customerSessionFromRequest(r)
	if err != nil {
		writeAppError(w, err)
		return
	}
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, models.BadRequest("invalid_json", "invalid JSON body"))
		return
	}
	customer := models.CustomerUser{ID: session.UserID, Email: session.Email, Name: session.Name}
	order, err := api.service.CreateOrder(req, customer)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (api *API) customerRegisterStart(w http.ResponseWriter, r *http.Request) {
	var req models.CustomerRegistrationStartRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	response, err := api.service.StartCustomerRegistration(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (api *API) customerRegisterVerify(w http.ResponseWriter, r *http.Request) {
	var req models.CustomerRegistrationVerifyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	response, session, expiresAt, err := api.service.VerifyCustomerRegistration(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	setCustomerSessionCookie(w, session.ID, expiresAt)
	writeJSON(w, http.StatusOK, response)
}

func (api *API) customerPasswordResetStart(w http.ResponseWriter, r *http.Request) {
	var req models.CustomerPasswordResetStartRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	response, err := api.service.StartCustomerPasswordReset(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (api *API) customerPasswordResetVerify(w http.ResponseWriter, r *http.Request) {
	var req models.CustomerPasswordResetVerifyRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	response, session, expiresAt, err := api.service.VerifyCustomerPasswordReset(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	setCustomerSessionCookie(w, session.ID, expiresAt)
	writeJSON(w, http.StatusOK, response)
}

func (api *API) customerLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	response, session, expiresAt, err := api.service.CustomerLogin(req)
	if err != nil {
		writeAppError(w, err)
		return
	}
	setCustomerSessionCookie(w, session.ID, expiresAt)
	writeJSON(w, http.StatusOK, response)
}

func (api *API) customerSession(w http.ResponseWriter, r *http.Request) {
	session, err := api.customerSessionFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusOK, models.CustomerSessionResponse{Authenticated: false})
		return
	}
	writeJSON(w, http.StatusOK, models.CustomerSessionResponse{Authenticated: true, Email: session.Email, Name: session.Name})
}

func (api *API) customerLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(customerSessionCookie)
	if err == nil && cookie.Value != "" {
		_ = api.service.CustomerLogout(cookie.Value)
	}
	clearCustomerSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) customerOrders(w http.ResponseWriter, r *http.Request) {
	session, err := api.customerSessionFromRequest(r)
	if err != nil {
		writeAppError(w, err)
		return
	}
	orders, err := api.service.CustomerOrders(session.UserID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (api *API) customerOrder(w http.ResponseWriter, r *http.Request) {
	session, err := api.customerSessionFromRequest(r)
	if err != nil {
		writeAppError(w, err)
		return
	}
	order, err := api.service.CustomerOrder(r.PathValue("orderID"), session.UserID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (api *API) adminCategories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Categories())
}

func (api *API) createCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if !decodeJSON(w, r, &category) {
		return
	}
	created, err := api.service.CreateCategory(category)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
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
	created, err := api.service.CreateProduct(product)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
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

func (api *API) uploadProductImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeAppError(w, models.BadRequest("invalid_multipart", "invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAppError(w, models.BadRequest("file_required", "file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	product, err := api.service.UploadProductImage(r.Context(), r.PathValue("productID"), header.Filename, contentType, file, header.Size)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (api *API) uploadProductAdditionalImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeAppError(w, models.BadRequest("invalid_multipart", "invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAppError(w, models.BadRequest("file_required", "file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	product, err := api.service.UploadProductAdditionalImage(r.Context(), r.PathValue("productID"), header.Filename, contentType, file, header.Size)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (api *API) deleteProductImage(w http.ResponseWriter, r *http.Request) {
	imageID, ok := parseID(w, r.PathValue("imageID"))
	if !ok {
		return
	}
	if err := api.service.DeleteProductImage(r.PathValue("productID"), imageID); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := api.service.DeleteProduct(r.PathValue("productID")); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) adminBlog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Blog())
}

func (api *API) createBlogPost(w http.ResponseWriter, r *http.Request) {
	var post models.BlogPost
	if !decodeJSON(w, r, &post) {
		return
	}
	created, err := api.service.CreateBlogPost(post)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api *API) updateBlogPost(w http.ResponseWriter, r *http.Request) {
	var post models.BlogPost
	if !decodeJSON(w, r, &post) {
		return
	}
	updated, err := api.service.UpdateBlogPost(r.PathValue("postID"), post)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (api *API) uploadBlogPostImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeAppError(w, models.BadRequest("invalid_multipart", "invalid multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAppError(w, models.BadRequest("file_required", "file is required"))
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	post, err := api.service.UploadBlogPostImage(r.Context(), r.PathValue("postID"), header.Filename, contentType, file, header.Size)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (api *API) deleteBlogPost(w http.ResponseWriter, r *http.Request) {
	if err := api.service.DeleteBlogPost(r.PathValue("postID")); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) adminGallery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.service.Gallery())
}

func (api *API) createGalleryItem(w http.ResponseWriter, r *http.Request) {
	var item models.GalleryItem
	if !decodeJSON(w, r, &item) {
		return
	}
	created, err := api.service.CreateGalleryItem(item)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api *API) updateGalleryItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("galleryItemID"))
	if !ok {
		return
	}
	var item models.GalleryItem
	if !decodeJSON(w, r, &item) {
		return
	}
	updated, err := api.service.UpdateGalleryItem(id, item)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (api *API) uploadGalleryItemImage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("galleryItemID"))
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeAppError(w, models.BadRequest("invalid_multipart", "invalid multipart form"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeAppError(w, models.BadRequest("file_required", "file is required"))
		return
	}
	defer file.Close()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	item, err := api.service.UploadGalleryItemImage(r.Context(), id, header.Filename, contentType, file, header.Size)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (api *API) deleteGalleryItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r.PathValue("galleryItemID"))
	if !ok {
		return
	}
	if err := api.service.DeleteGalleryItem(id); err != nil {
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

func (api *API) customer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := api.customerSessionFromRequest(r); err != nil {
			writeAppError(w, err)
			return
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

func (api *API) customerSessionFromRequest(r *http.Request) (models.CustomerSession, error) {
	cookie, err := r.Cookie(customerSessionCookie)
	if err != nil {
		return models.CustomerSession{}, models.Unauthorized("session_required", "customer session is required")
	}
	return api.service.CustomerSession(cookie.Value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeAppError(w, models.BadRequest("invalid_json", "invalid JSON body"))
		return false
	}
	return true
}

func parseID(w http.ResponseWriter, value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeAppError(w, models.BadRequest("id_invalid", "id is invalid"))
		return 0, false
	}
	return id, true
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

func setCustomerSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     customerSessionCookie,
		Value:    sessionID,
		Path:     "/api",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCustomerSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     customerSessionCookie,
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
