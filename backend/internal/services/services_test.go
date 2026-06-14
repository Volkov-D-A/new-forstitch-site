package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/repository"
	"new-forstitch-site/backend/internal/testutil"
)

func TestValidationHelpers(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "empty order", err: validateOrder(models.OrderRequest{}), code: "order_empty"},
		{
			name: "missing product id",
			err:  validateOrder(models.OrderRequest{Items: []models.CartItem{{Quantity: 1}}}),
			code: "product_id_required",
		},
		{
			name: "invalid quantity",
			err:  validateOrder(models.OrderRequest{Items: []models.CartItem{{ProductID: "product", Quantity: 0}}}),
			code: "quantity_invalid",
		},
		{name: "category label", err: validateCategory(models.Category{ID: "category"}), code: "label_required"},
		{
			name: "product title",
			err:  validateProduct(models.Product{ID: "product", Cat: "category"}),
			code: "title_required",
		},
		{
			name: "product category",
			err:  validateProduct(models.Product{ID: "product", Title: "Product"}),
			code: "category_required",
		},
		{
			name: "product price",
			err:  validateProduct(models.Product{ID: "product", Title: "Product", Cat: "category", Price: -1}),
			code: "price_invalid",
		},
		{name: "gallery title", err: validateGalleryItem(models.GalleryItem{Description: "Text"}), code: "title_required"},
		{name: "gallery description", err: validateGalleryItem(models.GalleryItem{Title: "Title"}), code: "description_required"},
		{
			name: "blog date format",
			err: validateBlogPost(models.BlogPost{
				ID: "post", Title: "Post", Date: "14.06.2026", Content: "Text",
			}),
			code: "date_invalid",
		},
		{name: "testimonial name", err: validateTestimonial(models.Testimonial{Text: "Text"}), code: "name_required"},
		{name: "testimonial text", err: validateTestimonial(models.Testimonial{Name: "Name"}), code: "text_required"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertErrorCode(t, test.err, test.code)
		})
	}

	if err := validateOrder(models.OrderRequest{
		Items: []models.CartItem{{ProductID: "product", Quantity: 1}},
	}); err != nil {
		t.Fatalf("expected valid order: %v", err)
	}
}

func TestBlogTextNormalization(t *testing.T) {
	content := `{"type":"doc","content":[{"type":"heading","content":[{"type":"text","text":"Заголовок"}]},{"type":"paragraph","content":[{"type":"text","text":"Первый абзац."}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Пункт списка"}]}]}]}]}`
	if got := blogPlainText(content); got != "Заголовок Первый абзац. Пункт списка" {
		t.Fatalf("unexpected rich text conversion: %q", got)
	}
	if got := blogPlainText("  plain \n text  "); got != "plain text" {
		t.Fatalf("unexpected plain text conversion: %q", got)
	}

	long := strings.Repeat("слово ", 60)
	excerpt := blogExcerpt(long)
	if len([]rune(excerpt)) > 243 || !strings.HasSuffix(excerpt, "...") {
		t.Fatalf("expected shortened excerpt, got %q", excerpt)
	}
}

func TestContentAndOrderLifecycle(t *testing.T) {
	service := New(testutil.NewRepositoryMock())

	category, err := service.CreateCategory(models.Category{Label: " Новая категория "})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if category.ID == "" {
		t.Fatal("expected generated category id")
	}

	product, err := service.CreateProduct(models.Product{
		Title: "Новая схема", Cat: category.ID, Price: 350,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	if product.ID == "" {
		t.Fatal("expected generated product id")
	}
	if _, err := service.Product(product.ID); err != nil {
		t.Fatalf("get product: %v", err)
	}

	gallery, err := service.CreateGalleryItem(models.GalleryItem{
		Title: "  Работа  ", Description: "  Готовый отшив  ",
	})
	if err != nil {
		t.Fatalf("create gallery item: %v", err)
	}
	if gallery.Title != "Работа" || gallery.Description != "Готовый отшив" {
		t.Fatalf("gallery item was not normalized: %+v", gallery)
	}

	post, err := service.CreateBlogPost(models.BlogPost{
		Title: " Новость ", Date: "2026-06-14", Content: "  Полный текст записи.  ",
	})
	if err != nil {
		t.Fatalf("create blog post: %v", err)
	}
	if post.Title != "Новость" || post.Excerpt != "Полный текст записи." {
		t.Fatalf("blog post was not normalized: %+v", post)
	}

	testimonial, err := service.CreateTestimonial(models.Testimonial{
		Name: " Анна ", Text: " Очень удобно. ",
	})
	if err != nil {
		t.Fatalf("create testimonial: %v", err)
	}
	if testimonial.Name != "Анна" || testimonial.Text != "Очень удобно." {
		t.Fatalf("testimonial was not normalized: %+v", testimonial)
	}

	settings, err := service.UpdateSiteSettings(models.SiteSettings{FeaturedProductID: product.ID})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if settings.FeaturedProductID != product.ID {
		t.Fatalf("unexpected settings: %+v", settings)
	}

	order, err := service.CreateOrder(models.OrderRequest{
		Items: []models.CartItem{{ProductID: product.ID, Quantity: 2}},
	}, models.CustomerUser{Email: "buyer@example.com", Name: "Анна"})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.ID == "" || order.Status != "paid" {
		t.Fatalf("unexpected order response: %+v", order)
	}
}

func TestAdminAuthenticationLifecycle(t *testing.T) {
	service := New(testutil.NewRepositoryMock())

	assertErrorCode(t, service.EnsureAdminUser("", "password"), "username_required")
	assertErrorCode(t, service.EnsureAdminUser("admin", ""), "password_required")
	if err := service.EnsureAdminUser("admin", "password"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}

	assertLoginErrorCode(t, service, models.LoginRequest{Username: "admin", Password: "wrong"}, "invalid_credentials")
	response, session, expiresAt, err := service.Login(models.LoginRequest{
		Username: " admin ", Password: "password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if response.Username != "admin" || response.CSRFToken == "" || session.ID == "" {
		t.Fatalf("unexpected login result: response=%+v session=%+v", response, session)
	}
	if !expiresAt.After(time.Now().Add(AdminSessionTTL - time.Minute)) {
		t.Fatalf("unexpected session expiration: %v", expiresAt)
	}
	if _, err := service.Session(session.ID); err != nil {
		t.Fatalf("get session: %v", err)
	}
	assertErrorCode(t, service.CheckCSRF(session, "wrong"), "csrf_invalid")
	if err := service.CheckCSRF(session, session.CSRFToken); err != nil {
		t.Fatalf("check csrf: %v", err)
	}
	if err := service.Logout(session.ID); err != nil {
		t.Fatalf("logout: %v", err)
	}
	assertErrorCode(t, mustSessionError(service, session.ID), "session_invalid")
}

func TestCustomerRegistrationAndPasswordReset(t *testing.T) {
	repo := testutil.NewRepositoryMock()
	service := New(repo)
	sender := &capturingMailer{}
	service.ConfigureMailer(sender, "https://example.com/")

	assertRegistrationErrorCode(t, service, models.CustomerRegistrationStartRequest{
		Email: "invalid", Password: "secret123",
	}, "email_invalid")
	assertRegistrationErrorCode(t, service, models.CustomerRegistrationStartRequest{
		Email: "buyer@example.com", Password: "short",
	}, "password_short")

	start, err := service.StartCustomerRegistration(models.CustomerRegistrationStartRequest{
		Email: " Buyer@Example.com ", Name: "Анна", Password: "secret123",
	})
	if err != nil {
		t.Fatalf("start registration: %v", err)
	}
	if start.Email != "buyer@example.com" {
		t.Fatalf("expected normalized email, got %q", start.Email)
	}
	registrationCode := sender.lastCode(t)

	_, _, _, err = service.VerifyCustomerRegistration(models.CustomerRegistrationVerifyRequest{
		Email: start.Email, Code: "000000",
	})
	assertErrorCode(t, err, "code_invalid")

	response, session, expiresAt, err := service.VerifyCustomerRegistration(models.CustomerRegistrationVerifyRequest{
		Email: start.Email, Code: registrationCode,
	})
	if err != nil {
		t.Fatalf("verify registration: %v", err)
	}
	if !response.Authenticated || session.Email != start.Email || !expiresAt.After(time.Now()) {
		t.Fatalf("unexpected registration result: response=%+v session=%+v", response, session)
	}

	login, _, _, err := service.CustomerLogin(models.LoginRequest{
		Username: "BUYER@example.com", Password: "secret123",
	})
	if err != nil || !login.Authenticated {
		t.Fatalf("customer login failed: response=%+v err=%v", login, err)
	}

	reset, err := service.StartCustomerPasswordReset(models.CustomerPasswordResetStartRequest{
		Email: "buyer@example.com",
	})
	if err != nil {
		t.Fatalf("start password reset: %v", err)
	}
	if reset.Email != start.Email {
		t.Fatalf("unexpected reset response: %+v", reset)
	}
	resetCode := sender.lastCode(t)

	_, _, _, err = service.VerifyCustomerPasswordReset(models.CustomerPasswordResetVerifyRequest{
		Email: start.Email, Code: resetCode, NewPassword: "short",
	})
	assertErrorCode(t, err, "password_short")

	_, resetSession, _, err := service.VerifyCustomerPasswordReset(models.CustomerPasswordResetVerifyRequest{
		Email: start.Email, Code: resetCode, NewPassword: "new-secret",
	})
	if err != nil {
		t.Fatalf("verify password reset: %v", err)
	}
	if resetSession.Email != start.Email {
		t.Fatalf("unexpected reset session: %+v", resetSession)
	}
	if _, _, _, err := service.CustomerLogin(models.LoginRequest{
		Username: start.Email, Password: "new-secret",
	}); err != nil {
		t.Fatalf("login with new password: %v", err)
	}

	unknown, err := service.StartCustomerPasswordReset(models.CustomerPasswordResetStartRequest{
		Email: "unknown@example.com",
	})
	if err != nil || unknown.Email != "unknown@example.com" {
		t.Fatalf("password reset should not reveal unknown account: response=%+v err=%v", unknown, err)
	}
}

func TestFileStorageLifecycle(t *testing.T) {
	repo := testutil.NewRepositoryMock()
	files := newFakeFileStorage()
	service := New(repo)
	service.ConfigureFiles(files, "https://cdn.example.com/")
	ctx := context.Background()

	product, err := service.UploadProductImage(
		ctx, "lighthouse_aniva", "cover.JPG", "image/jpeg", strings.NewReader("image"), 5,
	)
	if err != nil {
		t.Fatalf("upload product image: %v", err)
	}
	if product.Img != "https://cdn.example.com/products/lighthouse_aniva/cover.JPG" {
		t.Fatalf("unexpected product image URL: %q", product.Img)
	}

	product, err = service.UploadProductAdditionalImage(
		ctx, "lighthouse_aniva", "detail.png", "image/png", strings.NewReader("detail"), 6,
	)
	if err != nil {
		t.Fatalf("upload additional image: %v", err)
	}
	if len(product.Images) != 1 {
		t.Fatalf("expected one additional image, got %+v", product.Images)
	}

	product, err = service.UploadProductFile(
		ctx, "lighthouse_aniva", "scheme.pdf", "application/pdf", strings.NewReader("pdf"), 3,
	)
	if err != nil {
		t.Fatalf("upload product file: %v", err)
	}
	if len(product.Files) != 1 || product.Files[0].Name != "scheme.pdf" {
		t.Fatalf("unexpected product files: %+v", product.Files)
	}

	reader, info, err := service.File(ctx, "/products/lighthouse_aniva/cover.JPG/")
	if err != nil {
		t.Fatalf("get file: %v", err)
	}
	defer reader.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(body) != "image" || info.ContentType != "image/jpeg" {
		t.Fatalf("unexpected file: body=%q info=%+v", body, info)
	}

	invalidUploads := []struct {
		name        string
		contentType string
		reader      io.Reader
		size        int64
		code        string
	}{
		{name: "missing reader", contentType: "image/png", size: 1, code: "file_required"},
		{name: "empty file", contentType: "image/png", reader: strings.NewReader(""), code: "file_empty"},
		{name: "wrong type", contentType: "application/pdf", reader: strings.NewReader("x"), size: 1, code: "file_type_invalid"},
		{name: "too large", contentType: "image/png", reader: strings.NewReader("x"), size: maxProductImageSize + 1, code: "file_too_large"},
	}
	for _, test := range invalidUploads {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.UploadProductImage(
				ctx, "lighthouse_aniva", "image.png", test.contentType, test.reader, test.size,
			)
			assertErrorCode(t, err, test.code)
		})
	}
}

func TestServiceUpdateAndDeleteOperations(t *testing.T) {
	var calls []string
	repo := &repositoryStub{
		updateCategory: func(id string, category models.Category) error {
			calls = append(calls, "update-category:"+id+":"+category.ID)
			return nil
		},
		deleteCategory: func(id string) error {
			calls = append(calls, "delete-category:"+id)
			return nil
		},
		updateProduct: func(id string, product models.Product) error {
			calls = append(calls, "update-product:"+id+":"+product.ID)
			return nil
		},
		deleteProductImage: func(id string, imageID int64) error {
			calls = append(calls, "delete-image:"+id)
			return nil
		},
		deleteProductFile: func(id string, fileID int64) error {
			calls = append(calls, "delete-file:"+id)
			return nil
		},
		deleteProduct: func(id string) error {
			calls = append(calls, "delete-product:"+id)
			return nil
		},
		updateGalleryItem: func(id int64, item models.GalleryItem) error {
			calls = append(calls, "update-gallery")
			return nil
		},
		deleteGalleryItem: func(id int64) error {
			calls = append(calls, "delete-gallery")
			return nil
		},
		updateBlogPost: func(id string, post models.BlogPost) error {
			calls = append(calls, "update-blog:"+id)
			return nil
		},
		deleteBlogPost: func(id string) error {
			calls = append(calls, "delete-blog:"+id)
			return nil
		},
		updateTestimonial: func(id int64, testimonial models.Testimonial) error {
			calls = append(calls, "update-testimonial")
			return nil
		},
		deleteTestimonial: func(id int64) error {
			calls = append(calls, "delete-testimonial")
			return nil
		},
		orders: func() ([]models.Order, error) {
			return []models.Order{{ID: "order-1"}}, nil
		},
		orderForCustomer: func(orderID string, customerID int64) (models.Order, error) {
			return models.Order{ID: orderID, CustomerName: "Customer"}, nil
		},
	}
	service := New(repo)

	if err := service.UpdateCategory("category", models.Category{Label: "Category"}); err != nil {
		t.Fatalf("update category: %v", err)
	}
	if err := service.DeleteCategory("category"); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	if err := service.UpdateProduct("product", models.Product{Title: "Product", Cat: "category"}); err != nil {
		t.Fatalf("update product: %v", err)
	}
	if err := service.DeleteProductImage("product", 1); err != nil {
		t.Fatalf("delete product image: %v", err)
	}
	if err := service.DeleteProductFile("product", 1); err != nil {
		t.Fatalf("delete product file: %v", err)
	}
	if err := service.DeleteProduct("product"); err != nil {
		t.Fatalf("delete product: %v", err)
	}
	gallery, err := service.UpdateGalleryItem(1, models.GalleryItem{Title: " Title ", Description: " Description "})
	if err != nil || gallery.ID != 1 || gallery.Title != "Title" {
		t.Fatalf("update gallery item: item=%+v err=%v", gallery, err)
	}
	if err := service.DeleteGalleryItem(1); err != nil {
		t.Fatalf("delete gallery item: %v", err)
	}
	post, err := service.UpdateBlogPost("post", models.BlogPost{
		Title: "Post", Date: "2026-06-14", Content: "Content",
	})
	if err != nil || post.ID != "post" {
		t.Fatalf("update blog post: post=%+v err=%v", post, err)
	}
	if err := service.DeleteBlogPost("post"); err != nil {
		t.Fatalf("delete blog post: %v", err)
	}
	testimonial, err := service.UpdateTestimonial(1, models.Testimonial{Name: " Name ", Text: " Text "})
	if err != nil || testimonial.ID != 1 || testimonial.Name != "Name" {
		t.Fatalf("update testimonial: testimonial=%+v err=%v", testimonial, err)
	}
	if err := service.DeleteTestimonial(1); err != nil {
		t.Fatalf("delete testimonial: %v", err)
	}
	orders, err := service.AdminOrders()
	if err != nil || len(orders) != 1 {
		t.Fatalf("admin orders: orders=%+v err=%v", orders, err)
	}
	order, err := service.CustomerOrder("order-1", 1)
	if err != nil || order.ID != "order-1" {
		t.Fatalf("customer order: order=%+v err=%v", order, err)
	}

	if len(calls) != 12 {
		t.Fatalf("expected 12 repository calls, got %d: %v", len(calls), calls)
	}

	invalid := []struct {
		name string
		err  error
		code string
	}{
		{name: "update category", err: service.UpdateCategory("", models.Category{Label: "Category"}), code: "id_required"},
		{name: "delete category", err: service.DeleteCategory(" "), code: "id_required"},
		{name: "update product", err: service.UpdateProduct("", models.Product{Title: "Product", Cat: "category"}), code: "id_required"},
		{name: "delete image product", err: service.DeleteProductImage("", 1), code: "id_required"},
		{name: "delete image id", err: service.DeleteProductImage("product", 0), code: "image_id_invalid"},
		{name: "delete file product", err: service.DeleteProductFile("", 1), code: "id_required"},
		{name: "delete file id", err: service.DeleteProductFile("product", 0), code: "file_id_invalid"},
		{name: "delete product", err: service.DeleteProduct(""), code: "id_required"},
		{name: "delete gallery", err: service.DeleteGalleryItem(0), code: "id_required"},
		{name: "delete blog", err: service.DeleteBlogPost(""), code: "id_required"},
		{name: "delete testimonial", err: service.DeleteTestimonial(0), code: "id_required"},
		{name: "customer order", err: mustCustomerOrderError(service, "", 1), code: "order_id_required"},
	}
	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			assertErrorCode(t, test.err, test.code)
		})
	}
}

func TestContentImageUploadsAndCustomerFile(t *testing.T) {
	files := newFakeFileStorage()
	repo := &repositoryStub{}
	repo.updateGalleryItemImage = func(id int64, imageURL string) error {
		repo.galleryItems = []models.GalleryItem{{ID: id, Img: imageURL, Title: "Gallery"}}
		return nil
	}
	repo.updateBlogPostImage = func(id string, imageURL string) error {
		repo.blogPosts = []models.BlogPost{{ID: id, Img: imageURL, Title: "Post"}}
		return nil
	}
	repo.updateTestimonialImage = func(id int64, imageURL string) error {
		repo.testimonials = []models.Testimonial{{ID: id, Img: imageURL, Name: "Name"}}
		return nil
	}
	repo.productFileForCustomerOrder = func(orderID string, customerID int64, fileID int64) (models.ProductFile, error) {
		return models.ProductFile{ID: fileID, Name: "scheme.pdf", ObjectName: "downloads/scheme.pdf"}, nil
	}
	files.files["downloads/scheme.pdf"] = fakeStoredFile{data: []byte("pdf"), contentType: "application/pdf"}

	service := New(repo)
	service.ConfigureFiles(files, "https://cdn.example.com")
	ctx := context.Background()

	gallery, err := service.UploadGalleryItemImage(ctx, 1, "gallery.jpg", "image/jpeg", strings.NewReader("g"), 1)
	if err != nil || gallery.Img != "https://cdn.example.com/gallery/1/gallery.jpg" {
		t.Fatalf("upload gallery image: item=%+v err=%v", gallery, err)
	}
	post, err := service.UploadBlogPostImage(ctx, "post", "post.jpg", "image/jpeg", strings.NewReader("b"), 1)
	if err != nil || post.Img != "https://cdn.example.com/blog/post/post.jpg" {
		t.Fatalf("upload blog image: post=%+v err=%v", post, err)
	}
	contentURL, err := service.UploadBlogContentImage(ctx, "content.png", "image/png", strings.NewReader("c"), 1)
	if err != nil || contentURL != "https://cdn.example.com/blog/content/content.png" {
		t.Fatalf("upload content image: url=%q err=%v", contentURL, err)
	}
	testimonial, err := service.UploadTestimonialImage(ctx, 1, "avatar.png", "image/png", strings.NewReader("t"), 1)
	if err != nil || testimonial.Img != "https://cdn.example.com/testimonials/1/avatar.png" {
		t.Fatalf("upload testimonial image: testimonial=%+v err=%v", testimonial, err)
	}

	reader, info, err := service.CustomerOrderFile(ctx, "order-1", 1, 1)
	if err != nil {
		t.Fatalf("get customer order file: %v", err)
	}
	defer reader.Close()
	if info.Name != "scheme.pdf" || info.ContentType != "application/pdf" {
		t.Fatalf("unexpected file info: %+v", info)
	}

	_, _, err = service.CustomerOrderFile(ctx, "", 1, 1)
	assertErrorCode(t, err, "file_not_found")
}

func TestFileStorageErrorsAndLimits(t *testing.T) {
	ctx := context.Background()
	service := New(testutil.NewRepositoryMock())

	_, err := service.UploadProductImage(
		ctx, "lighthouse_aniva", "cover.jpg", "image/jpeg", strings.NewReader("x"), 1,
	)
	assertErrorCode(t, err, "file_storage_not_configured")
	_, err = service.UploadProductFile(
		ctx, "lighthouse_aniva", "scheme.pdf", "application/pdf", strings.NewReader("x"), 1,
	)
	assertErrorCode(t, err, "file_storage_not_configured")
	_, _, err = service.File(ctx, "")
	assertErrorCode(t, err, "file_not_found")
	_, _, err = service.CustomerOrderFile(ctx, "order", 1, 1)
	assertErrorCode(t, err, "file_not_found")

	files := &failingFileStorage{err: errors.New("storage unavailable")}
	service.ConfigureFiles(files, "https://cdn.example.com")
	_, err = service.UploadProductImage(
		ctx, "lighthouse_aniva", "cover.jpg", "image/jpeg", strings.NewReader("x"), 1,
	)
	if !errors.Is(err, files.err) {
		t.Fatalf("expected storage error, got %v", err)
	}
	_, _, err = service.File(ctx, "missing")
	assertErrorCode(t, err, "file_not_found")

	invalidFiles := []struct {
		name   string
		reader io.Reader
		size   int64
		code   string
	}{
		{name: "empty", reader: strings.NewReader(""), size: 0, code: "file_empty"},
		{name: "missing", reader: nil, size: 1, code: "file_empty"},
		{name: "too large", reader: strings.NewReader("x"), size: maxProductFileSize + 1, code: "file_too_large"},
	}
	for _, test := range invalidFiles {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.UploadProductFile(
				ctx, "lighthouse_aniva", "scheme.pdf", "application/pdf", test.reader, test.size,
			)
			assertErrorCode(t, err, test.code)
		})
	}
}

func TestAuthenticationValidationEdges(t *testing.T) {
	repo := testutil.NewRepositoryMock()
	service := New(repo)

	assertLoginErrorCode(t, service, models.LoginRequest{}, "credentials_required")
	_, _, _, err := service.CustomerLogin(models.LoginRequest{})
	assertErrorCode(t, err, "credentials_required")
	_, _, _, err = service.CustomerLogin(models.LoginRequest{
		Username: "missing@example.com", Password: "secret123",
	})
	assertErrorCode(t, err, "invalid_credentials")
	assertErrorCode(t, mustSessionError(service, ""), "session_required")
	_, err = service.CustomerSession("")
	assertErrorCode(t, err, "session_required")
	if err := service.Logout(""); err != nil {
		t.Fatalf("empty admin logout: %v", err)
	}
	if err := service.CustomerLogout(""); err != nil {
		t.Fatalf("empty customer logout: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, _, err := repo.EnsureCustomer("buyer@example.com", "Анна", string(hash)); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	_, err = service.StartCustomerRegistration(models.CustomerRegistrationStartRequest{
		Email: "buyer@example.com", Password: "secret123",
	})
	assertErrorCode(t, err, "customer_exists")
	_, err = service.StartCustomerPasswordReset(models.CustomerPasswordResetStartRequest{Email: "invalid"})
	assertErrorCode(t, err, "email_invalid")
	_, _, _, err = service.VerifyCustomerRegistration(models.CustomerRegistrationVerifyRequest{})
	assertErrorCode(t, err, "code_required")
	_, _, _, err = service.VerifyCustomerPasswordReset(models.CustomerPasswordResetVerifyRequest{})
	assertErrorCode(t, err, "code_required")

	service.ConfigureMailer(nil, "https://example.com/")
	if service.appBaseURL != "https://example.com" {
		t.Fatalf("expected trimmed app base URL, got %q", service.appBaseURL)
	}
}

type repositoryStub struct {
	repository.Repository

	updateCategory              func(string, models.Category) error
	deleteCategory              func(string) error
	updateProduct               func(string, models.Product) error
	deleteProductImage          func(string, int64) error
	deleteProductFile           func(string, int64) error
	productFileForCustomerOrder func(string, int64, int64) (models.ProductFile, error)
	deleteProduct               func(string) error
	updateGalleryItem           func(int64, models.GalleryItem) error
	updateGalleryItemImage      func(int64, string) error
	deleteGalleryItem           func(int64) error
	updateBlogPost              func(string, models.BlogPost) error
	updateBlogPostImage         func(string, string) error
	deleteBlogPost              func(string) error
	updateTestimonial           func(int64, models.Testimonial) error
	updateTestimonialImage      func(int64, string) error
	deleteTestimonial           func(int64) error
	orders                      func() ([]models.Order, error)
	orderForCustomer            func(string, int64) (models.Order, error)
	galleryItems                []models.GalleryItem
	blogPosts                   []models.BlogPost
	testimonials                []models.Testimonial
}

func (r *repositoryStub) UpdateCategory(id string, category models.Category) error {
	return r.updateCategory(id, category)
}

func (r *repositoryStub) DeleteCategory(id string) error {
	return r.deleteCategory(id)
}

func (r *repositoryStub) UpdateProduct(id string, product models.Product) error {
	return r.updateProduct(id, product)
}

func (r *repositoryStub) DeleteProductImage(id string, imageID int64) error {
	return r.deleteProductImage(id, imageID)
}

func (r *repositoryStub) DeleteProductFile(id string, fileID int64) error {
	return r.deleteProductFile(id, fileID)
}

func (r *repositoryStub) ProductFileForCustomerOrder(orderID string, customerID int64, fileID int64) (models.ProductFile, error) {
	return r.productFileForCustomerOrder(orderID, customerID, fileID)
}

func (r *repositoryStub) DeleteProduct(id string) error {
	return r.deleteProduct(id)
}

func (r *repositoryStub) Gallery() []models.GalleryItem {
	return r.galleryItems
}

func (r *repositoryStub) UpdateGalleryItem(id int64, item models.GalleryItem) error {
	return r.updateGalleryItem(id, item)
}

func (r *repositoryStub) UpdateGalleryItemImage(id int64, imageURL string) error {
	return r.updateGalleryItemImage(id, imageURL)
}

func (r *repositoryStub) DeleteGalleryItem(id int64) error {
	return r.deleteGalleryItem(id)
}

func (r *repositoryStub) Blog() []models.BlogPost {
	return r.blogPosts
}

func (r *repositoryStub) UpdateBlogPost(id string, post models.BlogPost) error {
	return r.updateBlogPost(id, post)
}

func (r *repositoryStub) UpdateBlogPostImage(id string, imageURL string) error {
	return r.updateBlogPostImage(id, imageURL)
}

func (r *repositoryStub) DeleteBlogPost(id string) error {
	return r.deleteBlogPost(id)
}

func (r *repositoryStub) Testimonials() []models.Testimonial {
	return r.testimonials
}

func (r *repositoryStub) UpdateTestimonial(id int64, testimonial models.Testimonial) error {
	return r.updateTestimonial(id, testimonial)
}

func (r *repositoryStub) UpdateTestimonialImage(id int64, imageURL string) error {
	return r.updateTestimonialImage(id, imageURL)
}

func (r *repositoryStub) DeleteTestimonial(id int64) error {
	return r.deleteTestimonial(id)
}

func (r *repositoryStub) Orders() ([]models.Order, error) {
	return r.orders()
}

func (r *repositoryStub) OrderForCustomer(orderID string, customerID int64) (models.Order, error) {
	return r.orderForCustomer(orderID, customerID)
}

func mustCustomerOrderError(service *Service, orderID string, customerID int64) error {
	_, err := service.CustomerOrder(orderID, customerID)
	return err
}

type capturingMailer struct {
	body string
}

func (m *capturingMailer) Send(_ string, _ string, body string) error {
	m.body = body
	return nil
}

func (m *capturingMailer) lastCode(t *testing.T) string {
	t.Helper()
	code := regexp.MustCompile(`\b\d{6}\b`).FindString(m.body)
	if code == "" {
		t.Fatalf("expected six-digit code in mail body %q", m.body)
	}
	return code
}

type fakeFileStorage struct {
	files map[string]fakeStoredFile
}

type fakeStoredFile struct {
	data        []byte
	contentType string
}

func newFakeFileStorage() *fakeFileStorage {
	return &fakeFileStorage{files: make(map[string]fakeStoredFile)}
}

func (s *fakeFileStorage) PutProductImage(_ context.Context, productID string, filename string, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("products/"+productID+"/"+filename, contentType, reader)
}

func (s *fakeFileStorage) PutProductFile(_ context.Context, productID string, filename string, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("product-files/"+productID+"/"+filename, contentType, reader)
}

func (s *fakeFileStorage) PutTestimonialImage(_ context.Context, testimonialID string, filename string, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("testimonials/"+testimonialID+"/"+filename, contentType, reader)
}

func (s *fakeFileStorage) PutBlogImage(_ context.Context, postID string, filename string, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("blog/"+postID+"/"+filename, contentType, reader)
}

func (s *fakeFileStorage) PutGalleryImage(_ context.Context, itemID string, filename string, contentType string, reader io.Reader, _ int64) (string, error) {
	return s.put("gallery/"+itemID+"/"+filename, contentType, reader)
}

func (s *fakeFileStorage) Get(_ context.Context, objectName string) (io.ReadCloser, models.FileObject, error) {
	file, ok := s.files[objectName]
	if !ok {
		return nil, models.FileObject{}, errors.New("file not found")
	}
	return io.NopCloser(bytes.NewReader(file.data)), models.FileObject{
		ContentType: file.contentType,
		Size:        int64(len(file.data)),
	}, nil
}

func (s *fakeFileStorage) put(objectName string, contentType string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.files[objectName] = fakeStoredFile{data: data, contentType: contentType}
	return objectName, nil
}

type failingFileStorage struct {
	err error
}

func (s *failingFileStorage) PutProductImage(context.Context, string, string, string, io.Reader, int64) (string, error) {
	return "", s.err
}

func (s *failingFileStorage) PutProductFile(context.Context, string, string, string, io.Reader, int64) (string, error) {
	return "", s.err
}

func (s *failingFileStorage) PutTestimonialImage(context.Context, string, string, string, io.Reader, int64) (string, error) {
	return "", s.err
}

func (s *failingFileStorage) PutBlogImage(context.Context, string, string, string, io.Reader, int64) (string, error) {
	return "", s.err
}

func (s *failingFileStorage) PutGalleryImage(context.Context, string, string, string, io.Reader, int64) (string, error) {
	return "", s.err
}

func (s *failingFileStorage) Get(context.Context, string) (io.ReadCloser, models.FileObject, error) {
	return nil, models.FileObject{}, s.err
}

func assertLoginErrorCode(t *testing.T, service *Service, req models.LoginRequest, code string) {
	t.Helper()
	_, _, _, err := service.Login(req)
	assertErrorCode(t, err, code)
}

func assertRegistrationErrorCode(t *testing.T, service *Service, req models.CustomerRegistrationStartRequest, code string) {
	t.Helper()
	_, err := service.StartCustomerRegistration(req)
	assertErrorCode(t, err, code)
}

func mustSessionError(service *Service, sessionID string) error {
	_, err := service.Session(sessionID)
	return err
}

func assertErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %q", code)
	}
	var appErr models.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected error code %q, got %q", code, appErr.Code)
	}
}
