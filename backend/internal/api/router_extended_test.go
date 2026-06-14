package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/services"
	"new-forstitch-site/backend/internal/testutil"
)

func TestCustomerRegistrationAndPasswordResetEndpoints(t *testing.T) {
	repo := testutil.NewRepositoryMock()
	service := services.New(repo)
	sender := &apiCapturingMailer{}
	service.ConfigureMailer(sender, "http://localhost:3000")
	router := NewRouter(service, nil)

	startRec := serveJSON(router, http.MethodPost, "/api/customer/register/start",
		`{"email":"buyer@example.com","name":"Анна","password":"secret123"}`, nil)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start registration: %d %s", startRec.Code, startRec.Body.String())
	}
	registrationCode := sender.lastCode(t)

	verifyRec := serveJSON(router, http.MethodPost, "/api/customer/register/verify",
		`{"email":"buyer@example.com","code":"`+registrationCode+`"}`, nil)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify registration: %d %s", verifyRec.Code, verifyRec.Body.String())
	}
	assertResponseCookie(t, verifyRec, customerSessionCookie)

	resetStartRec := serveJSON(router, http.MethodPost, "/api/customer/password-reset/start",
		`{"email":"buyer@example.com"}`, nil)
	if resetStartRec.Code != http.StatusOK {
		t.Fatalf("start reset: %d %s", resetStartRec.Code, resetStartRec.Body.String())
	}
	resetCode := sender.lastCode(t)

	resetVerifyRec := serveJSON(router, http.MethodPost, "/api/customer/password-reset/verify",
		`{"email":"buyer@example.com","code":"`+resetCode+`","newPassword":"new-secret"}`, nil)
	if resetVerifyRec.Code != http.StatusOK {
		t.Fatalf("verify reset: %d %s", resetVerifyRec.Code, resetVerifyRec.Body.String())
	}
	assertResponseCookie(t, resetVerifyRec, customerSessionCookie)
}

func TestAdminCRUDHandlers(t *testing.T) {
	router, _, _ := testRouterWithFiles(t)
	adminCookie, csrfToken := loginAdmin(t, router)
	headers := map[string]string{"X-CSRF-Token": csrfToken}

	categoryRec := serveJSON(router, http.MethodPost, "/api/admin/categories",
		`{"label":"Удаляемая категория"}`, withCookie(headers, adminCookie))
	var category models.Category
	decodeResponse(t, categoryRec, http.StatusCreated, &category)

	updateCategoryRec := serveJSON(router, http.MethodPut, "/api/admin/categories/"+category.ID,
		`{"label":"Обновлённая категория"}`, withCookie(headers, adminCookie))
	if updateCategoryRec.Code != http.StatusOK ||
		!strings.Contains(updateCategoryRec.Body.String(), "Обновлённая категория") {
		t.Fatalf("update category: %d %s", updateCategoryRec.Code, updateCategoryRec.Body.String())
	}
	deleteCategoryRec := serveJSON(router, http.MethodDelete, "/api/admin/categories/"+category.ID, "",
		withCookie(headers, adminCookie))
	if deleteCategoryRec.Code != http.StatusNoContent {
		t.Fatalf("delete category: %d %s", deleteCategoryRec.Code, deleteCategoryRec.Body.String())
	}

	productRec := serveJSON(router, http.MethodPost, "/api/admin/products",
		`{"title":"Новая схема","cat":"landscape","price":500,"size":"100x100","colors":"20"}`,
		withCookie(headers, adminCookie))
	var product models.Product
	decodeResponse(t, productRec, http.StatusCreated, &product)

	updateProductRec := serveJSON(router, http.MethodPut, "/api/admin/products/"+product.ID,
		`{"title":"Обновлённая схема","cat":"landscape","price":550,"size":"100x100","colors":"20"}`,
		withCookie(headers, adminCookie))
	if updateProductRec.Code != http.StatusOK ||
		!strings.Contains(updateProductRec.Body.String(), "Обновлённая схема") {
		t.Fatalf("update product: %d %s", updateProductRec.Code, updateProductRec.Body.String())
	}

	testimonialRec := serveJSON(router, http.MethodPost, "/api/admin/testimonials",
		`{"name":"Анна","text":"Отличная схема"}`, withCookie(headers, adminCookie))
	var testimonial models.Testimonial
	decodeResponse(t, testimonialRec, http.StatusCreated, &testimonial)
	updateTestimonialRec := serveJSON(router, http.MethodPut, "/api/admin/testimonials/"+strconv.FormatInt(testimonial.ID, 10),
		`{"name":"Мария","text":"Обновлённый отзыв"}`, withCookie(headers, adminCookie))
	if updateTestimonialRec.Code != http.StatusOK {
		t.Fatalf("update testimonial: %d %s", updateTestimonialRec.Code, updateTestimonialRec.Body.String())
	}
	deleteTestimonialRec := serveJSON(router, http.MethodDelete, "/api/admin/testimonials/"+strconv.FormatInt(testimonial.ID, 10), "",
		withCookie(headers, adminCookie))
	if deleteTestimonialRec.Code != http.StatusNoContent {
		t.Fatalf("delete testimonial: %d %s", deleteTestimonialRec.Code, deleteTestimonialRec.Body.String())
	}

	blogRec := serveJSON(router, http.MethodPost, "/api/admin/blog",
		`{"title":"Запись","date":"2026-06-14","content":"Текст"}`, withCookie(headers, adminCookie))
	var post models.BlogPost
	decodeResponse(t, blogRec, http.StatusCreated, &post)
	updateBlogRec := serveJSON(router, http.MethodPut, "/api/admin/blog/"+post.ID,
		`{"title":"Обновлённая запись","date":"2026-06-14","content":"Новый текст"}`,
		withCookie(headers, adminCookie))
	if updateBlogRec.Code != http.StatusOK {
		t.Fatalf("update blog: %d %s", updateBlogRec.Code, updateBlogRec.Body.String())
	}
	deleteBlogRec := serveJSON(router, http.MethodDelete, "/api/admin/blog/"+post.ID, "",
		withCookie(headers, adminCookie))
	if deleteBlogRec.Code != http.StatusNoContent {
		t.Fatalf("delete blog: %d %s", deleteBlogRec.Code, deleteBlogRec.Body.String())
	}

	galleryRec := serveJSON(router, http.MethodPost, "/api/admin/gallery",
		`{"title":"Работа","description":"Описание"}`, withCookie(headers, adminCookie))
	var item models.GalleryItem
	decodeResponse(t, galleryRec, http.StatusCreated, &item)
	updateGalleryRec := serveJSON(router, http.MethodPut, "/api/admin/gallery/"+strconv.FormatInt(item.ID, 10),
		`{"title":"Обновлённая работа","description":"Новое описание"}`,
		withCookie(headers, adminCookie))
	if updateGalleryRec.Code != http.StatusOK {
		t.Fatalf("update gallery: %d %s", updateGalleryRec.Code, updateGalleryRec.Body.String())
	}
	deleteGalleryRec := serveJSON(router, http.MethodDelete, "/api/admin/gallery/"+strconv.FormatInt(item.ID, 10), "",
		withCookie(headers, adminCookie))
	if deleteGalleryRec.Code != http.StatusNoContent {
		t.Fatalf("delete gallery: %d %s", deleteGalleryRec.Code, deleteGalleryRec.Body.String())
	}

	deleteProductRec := serveJSON(router, http.MethodDelete, "/api/admin/products/"+product.ID, "",
		withCookie(headers, adminCookie))
	if deleteProductRec.Code != http.StatusNoContent {
		t.Fatalf("delete product: %d %s", deleteProductRec.Code, deleteProductRec.Body.String())
	}
}

func TestAdminMultipartHandlersAndFileDownloads(t *testing.T) {
	router, repo, files := testRouterWithFiles(t)
	adminCookie, csrfToken := loginAdmin(t, router)

	productImageRec := serveMultipart(t, router, "/api/admin/products/lighthouse_aniva/image",
		"cover.jpg", "image/jpeg", "image", adminCookie, csrfToken)
	if productImageRec.Code != http.StatusOK ||
		!strings.Contains(productImageRec.Body.String(), "/api/files/products/lighthouse_aniva/cover.jpg") {
		t.Fatalf("upload product image: %d %s", productImageRec.Code, productImageRec.Body.String())
	}
	publicFileRec := httptest.NewRecorder()
	router.ServeHTTP(publicFileRec, httptest.NewRequest(
		http.MethodGet, "/api/files/products/lighthouse_aniva/cover.jpg", nil,
	))
	if publicFileRec.Code != http.StatusOK || publicFileRec.Body.String() != "image" {
		t.Fatalf("download public file: %d %q", publicFileRec.Code, publicFileRec.Body.String())
	}

	additionalRec := serveMultipart(t, router, "/api/admin/products/lighthouse_aniva/images",
		"detail.png", "image/png", "detail", adminCookie, csrfToken)
	var product models.Product
	decodeResponse(t, additionalRec, http.StatusOK, &product)
	if len(product.Images) != 1 {
		t.Fatalf("expected additional image, got %+v", product.Images)
	}
	deleteImageRec := serveJSON(router, http.MethodDelete,
		"/api/admin/products/lighthouse_aniva/images/"+strconv.FormatInt(product.Images[0].ID, 10), "",
		withCookie(map[string]string{"X-CSRF-Token": csrfToken}, adminCookie))
	if deleteImageRec.Code != http.StatusNoContent {
		t.Fatalf("delete image: %d %s", deleteImageRec.Code, deleteImageRec.Body.String())
	}

	productFileRec := serveMultipart(t, router, "/api/admin/products/lighthouse_aniva/files",
		"scheme.pdf", "application/pdf", "pdf-content", adminCookie, csrfToken)
	decodeResponse(t, productFileRec, http.StatusOK, &product)
	if len(product.Files) != 1 {
		t.Fatalf("expected product file, got %+v", product.Files)
	}

	customerCookie := seedAndLoginCustomer(t, router, repo)
	createOrderRec := serveJSON(router, http.MethodPost, "/api/orders",
		`{"items":[{"productId":"lighthouse_aniva","quantity":1}]}`, withCookie(nil, customerCookie))
	var order models.OrderResponse
	decodeResponse(t, createOrderRec, http.StatusCreated, &order)

	downloadReq := httptest.NewRequest(http.MethodGet,
		"/api/customer/orders/"+order.ID+"/files/"+strconv.FormatInt(product.Files[0].ID, 10), nil)
	downloadReq.AddCookie(customerCookie)
	downloadRec := httptest.NewRecorder()
	router.ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusOK || downloadRec.Body.String() != "pdf-content" {
		t.Fatalf("download order file: %d %q", downloadRec.Code, downloadRec.Body.String())
	}
	if !strings.Contains(downloadRec.Header().Get("Content-Disposition"), "scheme.pdf") {
		t.Fatalf("unexpected content disposition: %q", downloadRec.Header().Get("Content-Disposition"))
	}

	testimonialImageRec := serveMultipart(t, router, "/api/admin/testimonials/1/image",
		"avatar.png", "image/png", "avatar", adminCookie, csrfToken)
	if testimonialImageRec.Code != http.StatusOK {
		t.Fatalf("upload testimonial image: %d %s", testimonialImageRec.Code, testimonialImageRec.Body.String())
	}
	blogImageRec := serveMultipart(t, router, "/api/admin/blog/new-patterns/image",
		"post.jpg", "image/jpeg", "post", adminCookie, csrfToken)
	if blogImageRec.Code != http.StatusOK {
		t.Fatalf("upload blog image: %d %s", blogImageRec.Code, blogImageRec.Body.String())
	}
	contentImageRec := serveMultipart(t, router, "/api/admin/blog/images",
		"content.jpg", "image/jpeg", "content", adminCookie, csrfToken)
	if contentImageRec.Code != http.StatusOK || !strings.Contains(contentImageRec.Body.String(), `"url":`) {
		t.Fatalf("upload content image: %d %s", contentImageRec.Code, contentImageRec.Body.String())
	}
	galleryImageRec := serveMultipart(t, router, "/api/admin/gallery/1/image",
		"gallery.jpg", "image/jpeg", "gallery", adminCookie, csrfToken)
	if galleryImageRec.Code != http.StatusOK {
		t.Fatalf("upload gallery image: %d %s", galleryImageRec.Code, galleryImageRec.Body.String())
	}

	deleteFileRec := serveJSON(router, http.MethodDelete,
		"/api/admin/products/lighthouse_aniva/files/"+strconv.FormatInt(product.Files[0].ID, 10), "",
		withCookie(map[string]string{"X-CSRF-Token": csrfToken}, adminCookie))
	if deleteFileRec.Code != http.StatusNoContent {
		t.Fatalf("delete file: %d %s", deleteFileRec.Code, deleteFileRec.Body.String())
	}

	if len(files.files) < 6 {
		t.Fatalf("expected uploaded files in fake storage, got %d", len(files.files))
	}
}

func TestMultipartValidationErrors(t *testing.T) {
	router, _, _ := testRouterWithFiles(t)
	cookie, csrfToken := loginAdmin(t, router)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/products/lighthouse_aniva/image",
		strings.NewReader("not multipart"))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Content-Type", "multipart/form-data")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest ||
		!strings.Contains(rec.Body.String(), `"code":"invalid_multipart"`) {
		t.Fatalf("expected invalid multipart response, got %d %s", rec.Code, rec.Body.String())
	}
}

func testRouterWithFiles(t *testing.T) (http.Handler, *testutil.RepositoryMock, *apiFileStorage) {
	t.Helper()
	repo := testutil.NewRepositoryMock()
	files := &apiFileStorage{files: make(map[string]apiStoredFile)}
	service := services.New(repo)
	service.ConfigureFiles(files, "/api/files")
	if err := service.EnsureAdminUser("admin", "password"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	return NewRouter(service, nil), repo, files
}

func seedAndLoginCustomer(t *testing.T, router http.Handler, repo *testutil.RepositoryMock) *http.Cookie {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, _, err := repo.EnsureCustomer("buyer@example.com", "Анна", string(hash)); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return loginCustomer(t, router)
}

func serveJSON(router http.Handler, method string, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for key, value := range headers {
		if strings.HasPrefix(key, "Cookie:") {
			req.AddCookie(&http.Cookie{Name: strings.TrimPrefix(key, "Cookie:"), Value: value})
			continue
		}
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func withCookie(headers map[string]string, cookie *http.Cookie) map[string]string {
	out := make(map[string]string, len(headers)+1)
	for key, value := range headers {
		out[key] = value
	}
	out["Cookie:"+cookie.Name] = cookie.Value
	return out
}

func serveMultipart(
	t *testing.T,
	router http.Handler,
	path string,
	filename string,
	contentType string,
	content string,
	cookie *http.Cookie,
	csrfToken string,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create multipart part: %v", err)
	}
	if _, err := io.WriteString(part, content); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, status int, target any) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, rec.Code, rec.Body.String())
	}
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertResponseCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name && cookie.Value != "" {
			return
		}
	}
	t.Fatalf("expected response cookie %q", name)
}

type apiCapturingMailer struct {
	body string
}

func (m *apiCapturingMailer) Send(_ string, _ string, body string) error {
	m.body = body
	return nil
}

func (m *apiCapturingMailer) lastCode(t *testing.T) string {
	t.Helper()
	code := regexp.MustCompile(`\b\d{6}\b`).FindString(m.body)
	if code == "" {
		t.Fatalf("expected mail code in %q", m.body)
	}
	return code
}

type apiFileStorage struct {
	files map[string]apiStoredFile
}

type apiStoredFile struct {
	data        []byte
	contentType string
}

func (s *apiFileStorage) PutProductImage(_ context.Context, productID, filename, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("products/"+productID+"/"+filename, contentType, reader)
}

func (s *apiFileStorage) PutProductFile(_ context.Context, productID, filename, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("product-files/"+productID+"/"+filename, contentType, reader)
}

func (s *apiFileStorage) PutTestimonialImage(_ context.Context, testimonialID, filename, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("testimonials/"+testimonialID+"/"+filename, contentType, reader)
}

func (s *apiFileStorage) PutBlogImage(_ context.Context, postID, filename, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("blog/"+postID+"/"+filename, contentType, reader)
}

func (s *apiFileStorage) PutGalleryImage(_ context.Context, itemID, filename, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("gallery/"+itemID+"/"+filename, contentType, reader)
}

func (s *apiFileStorage) Get(_ context.Context, objectName string) (io.ReadCloser, models.FileObject, error) {
	file, ok := s.files[objectName]
	if !ok {
		return nil, models.FileObject{}, errors.New("file not found")
	}
	return io.NopCloser(bytes.NewReader(file.data)), models.FileObject{
		ContentType: file.contentType,
		Size:        int64(len(file.data)),
	}, nil
}

func (s *apiFileStorage) put(objectName string, contentType string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.files[objectName] = apiStoredFile{data: data, contentType: contentType}
	return objectName, nil
}
