//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	appdb "new-forstitch-site/backend/internal/db"
	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/repository"
)

var integrationDB *sql.DB
var integrationRepo *repository.PostgresRepository

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "TEST_DATABASE_URL is required for integration tests")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	database, err := appdb.Open(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open integration database: %v\n", err)
		os.Exit(2)
	}
	defer database.Close()

	// Running migrations twice verifies that an already migrated database is safe.
	if err := appdb.Migrate(ctx, database); err != nil {
		fmt.Fprintf(os.Stderr, "migrate integration database: %v\n", err)
		os.Exit(2)
	}
	if err := appdb.Migrate(ctx, database); err != nil {
		fmt.Fprintf(os.Stderr, "repeat integration migrations: %v\n", err)
		os.Exit(2)
	}

	integrationDB = database
	integrationRepo = repository.NewPostgresRepository(database)
	code := m.Run()
	if err := database.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close integration database: %v\n", err)
		if code == 0 {
			code = 2
		}
	}
	os.Exit(code)
}

func TestMigrationsAndSeedData(t *testing.T) {
	var migrationCount int
	if err := integrationDB.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if migrationCount != 13 {
		t.Fatalf("expected 13 applied migrations, got %d", migrationCount)
	}

	categories := integrationRepo.Categories()
	if len(categories) < 5 {
		t.Fatalf("expected seeded categories, got %d", len(categories))
	}
	product, err := integrationRepo.Product("lighthouse_aniva")
	if err != nil {
		t.Fatalf("get seeded product: %v", err)
	}
	if product.Title != "Маяк на мысе Анива" || product.Cat != "landscape" {
		t.Fatalf("unexpected seeded product: %+v", product)
	}

	content := integrationRepo.SiteContent()
	if content.FeaturedProductID != "lighthouse_aniva" || len(content.HowToBuy) != 4 {
		t.Fatalf("unexpected seeded site content: %+v", content)
	}
}

func TestCatalogLifecycle(t *testing.T) {
	category := models.Category{ID: "integration-category", Label: "Integration Category"}
	if err := integrationRepo.CreateCategory(category); err != nil {
		t.Fatalf("create category: %v", err)
	}
	if err := integrationRepo.CreateCategory(category); err == nil {
		t.Fatal("expected duplicate category error")
	} else {
		assertAppError(t, err, models.ErrConflict, "record_exists")
	}

	category.Label = "Updated Category"
	if err := integrationRepo.UpdateCategory(category.ID, category); err != nil {
		t.Fatalf("update category: %v", err)
	}

	product := models.Product{
		ID:          "integration-product",
		Title:       "Integration Product",
		Price:       750,
		Cat:         category.ID,
		Img:         "https://example.com/product.jpg",
		Size:        "100 x 100",
		Colors:      "20",
		Description: "Integration description",
	}
	if err := integrationRepo.CreateProduct(product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	product.Title = "Updated Integration Product"
	product.Price = 800
	product.Img = "https://example.com/updated.jpg"
	if err := integrationRepo.UpdateProduct(product.ID, product); err != nil {
		t.Fatalf("update product: %v", err)
	}
	if err := integrationRepo.UpdateProductImage(product.ID, "https://example.com/cover.jpg"); err != nil {
		t.Fatalf("update product image: %v", err)
	}

	image, err := integrationRepo.AddProductImage(product.ID, "https://example.com/detail.jpg")
	if err != nil {
		t.Fatalf("add product image: %v", err)
	}
	file, err := integrationRepo.AddProductFile(product.ID, "scheme.pdf", "product-files/integration/scheme.pdf")
	if err != nil {
		t.Fatalf("add product file: %v", err)
	}

	stored, err := integrationRepo.Product(product.ID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}
	if len(stored.Images) != 1 || stored.Images[0].ID != image.ID ||
		len(stored.Files) != 1 || stored.Files[0].ID != file.ID {
		t.Fatalf("unexpected product media: %+v", stored)
	}
	if stored.Title != product.Title || stored.Price != product.Price ||
		stored.Img != "https://example.com/cover.jpg" {
		t.Fatalf("product update was not persisted: %+v", stored)
	}
	if len(integrationRepo.Products()) < 5 {
		t.Fatal("expected created product in products list")
	}

	if err := integrationRepo.DeleteCategory(category.ID); err == nil {
		t.Fatal("expected category reference error")
	} else {
		assertAppError(t, err, models.ErrValidation, "reference_not_found")
	}
	if err := integrationRepo.DeleteProductImage(product.ID, image.ID); err != nil {
		t.Fatalf("delete product image: %v", err)
	}
	if err := integrationRepo.DeleteProductFile(product.ID, file.ID); err != nil {
		t.Fatalf("delete product file: %v", err)
	}
	if err := integrationRepo.DeleteProduct(product.ID); err != nil {
		t.Fatalf("delete product: %v", err)
	}
	if err := integrationRepo.DeleteCategory(category.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}
}

func TestContentLifecycle(t *testing.T) {
	gallery, err := integrationRepo.CreateGalleryItem(models.GalleryItem{
		Title: "Integration Gallery", Description: "Gallery Description",
	})
	if err != nil {
		t.Fatalf("create gallery item: %v", err)
	}
	gallery.Title = "Updated Gallery"
	if err := integrationRepo.UpdateGalleryItem(gallery.ID, gallery); err != nil {
		t.Fatalf("update gallery item: %v", err)
	}
	if err := integrationRepo.UpdateGalleryItemImage(gallery.ID, "https://example.com/gallery.jpg"); err != nil {
		t.Fatalf("update gallery image: %v", err)
	}
	if !containsGallery(integrationRepo.Gallery(), gallery.ID, "Updated Gallery", "https://example.com/gallery.jpg") {
		t.Fatal("gallery changes were not persisted")
	}

	post := models.BlogPost{
		ID: "integration-post", Title: "Integration Post", Date: "2026-06-14",
		Tag: "Test", Excerpt: "Excerpt", Content: "Content",
	}
	post, err = integrationRepo.CreateBlogPost(post)
	if err != nil {
		t.Fatalf("create blog post: %v", err)
	}
	post.Title = "Updated Post"
	post.Content = "Updated Content"
	if err := integrationRepo.UpdateBlogPost(post.ID, post); err != nil {
		t.Fatalf("update blog post: %v", err)
	}
	if err := integrationRepo.UpdateBlogPostImage(post.ID, "https://example.com/post.jpg"); err != nil {
		t.Fatalf("update blog image: %v", err)
	}
	if !containsPost(integrationRepo.Blog(), post.ID, "Updated Post", "https://example.com/post.jpg") {
		t.Fatal("blog changes were not persisted")
	}

	testimonial, err := integrationRepo.CreateTestimonial(models.Testimonial{
		Name: "Integration User", Role: "Tester", Text: "Initial text",
	})
	if err != nil {
		t.Fatalf("create testimonial: %v", err)
	}
	testimonial.Name = "Updated User"
	testimonial.Text = "Updated text"
	if err := integrationRepo.UpdateTestimonial(testimonial.ID, testimonial); err != nil {
		t.Fatalf("update testimonial: %v", err)
	}
	if err := integrationRepo.UpdateTestimonialImage(testimonial.ID, "https://example.com/avatar.jpg"); err != nil {
		t.Fatalf("update testimonial image: %v", err)
	}
	if !containsTestimonial(
		integrationRepo.Testimonials(), testimonial.ID, "Updated User", "https://example.com/avatar.jpg",
	) {
		t.Fatal("testimonial changes were not persisted")
	}

	if err := integrationRepo.UpdateSiteSettings(models.SiteSettings{
		FeaturedProductID: "dragon_library",
	}); err != nil {
		t.Fatalf("update site settings: %v", err)
	}
	if got := integrationRepo.SiteContent().FeaturedProductID; got != "dragon_library" {
		t.Fatalf("unexpected featured product: %q", got)
	}

	if err := integrationRepo.DeleteGalleryItem(gallery.ID); err != nil {
		t.Fatalf("delete gallery item: %v", err)
	}
	if err := integrationRepo.DeleteBlogPost(post.ID); err != nil {
		t.Fatalf("delete blog post: %v", err)
	}
	if err := integrationRepo.DeleteTestimonial(testimonial.ID); err != nil {
		t.Fatalf("delete testimonial: %v", err)
	}
}

func TestAuthOrdersAndProtectedFiles(t *testing.T) {
	if err := integrationRepo.EnsureAdminUser("integration-admin", "hash-1"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	if err := integrationRepo.EnsureAdminUser("integration-admin", "hash-2"); err != nil {
		t.Fatalf("update admin: %v", err)
	}
	admin, err := integrationRepo.AdminUserByUsername("integration-admin")
	if err != nil || admin.PasswordHash != "hash-2" {
		t.Fatalf("unexpected admin: user=%+v err=%v", admin, err)
	}

	adminSession := models.AdminSession{
		ID: "integration-admin-session", UserID: admin.ID, Username: admin.Username, CSRFToken: "csrf",
	}
	if err := integrationRepo.CreateAdminSession(adminSession, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create admin session: %v", err)
	}
	if _, err := integrationRepo.AdminSession(adminSession.ID, time.Now()); err != nil {
		t.Fatalf("get admin session: %v", err)
	}
	expiredAdminSession := adminSession
	expiredAdminSession.ID = "expired-admin-session"
	if err := integrationRepo.CreateAdminSession(expiredAdminSession, time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("create expired admin session: %v", err)
	}
	if err := integrationRepo.DeleteExpiredAdminSessions(time.Now()); err != nil {
		t.Fatalf("delete expired admin sessions: %v", err)
	}
	if _, err := integrationRepo.AdminSession(expiredAdminSession.ID, time.Now()); err == nil {
		t.Fatal("expected expired admin session to be deleted")
	}
	if err := integrationRepo.DeleteAdminSession(adminSession.ID); err != nil {
		t.Fatalf("delete admin session: %v", err)
	}

	customer, created, err := integrationRepo.EnsureCustomer(
		"integration@example.com", "Integration Customer", "password-hash",
	)
	if err != nil || !created {
		t.Fatalf("create customer: customer=%+v created=%v err=%v", customer, created, err)
	}
	customer, created, err = integrationRepo.EnsureCustomer(
		customer.Email, "Updated Customer", "ignored-hash",
	)
	if err != nil || created || customer.Name != "Updated Customer" {
		t.Fatalf("update customer: customer=%+v created=%v err=%v", customer, created, err)
	}

	customerSession := models.CustomerSession{
		ID: "integration-customer-session", UserID: customer.ID, Email: customer.Email, Name: customer.Name,
	}
	if err := integrationRepo.CreateCustomerSession(customerSession, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create customer session: %v", err)
	}
	if _, err := integrationRepo.CustomerSession(customerSession.ID, time.Now()); err != nil {
		t.Fatalf("get customer session: %v", err)
	}
	expiredCustomerSession := customerSession
	expiredCustomerSession.ID = "expired-customer-session"
	if err := integrationRepo.CreateCustomerSession(expiredCustomerSession, time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("create expired customer session: %v", err)
	}
	if err := integrationRepo.DeleteExpiredCustomerSessions(time.Now()); err != nil {
		t.Fatalf("delete expired customer sessions: %v", err)
	}
	if _, err := integrationRepo.CustomerSession(expiredCustomerSession.ID, time.Now()); err == nil {
		t.Fatal("expected expired customer session to be deleted")
	}

	productFile, err := integrationRepo.AddProductFile(
		"lighthouse_aniva", "integration-scheme.pdf", "product-files/lighthouse_aniva/integration.pdf",
	)
	if err != nil {
		t.Fatalf("add order product file: %v", err)
	}
	defer integrationRepo.DeleteProductFile("lighthouse_aniva", productFile.ID)

	orderResponse, err := integrationRepo.CreateOrder(models.OrderRequest{
		Items: []models.CartItem{{ProductID: "lighthouse_aniva", Quantity: 2}},
	}, customer)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	order, err := integrationRepo.OrderForCustomer(orderResponse.ID, customer.ID)
	if err != nil {
		t.Fatalf("get customer order: %v", err)
	}
	if order.Status != "paid" || len(order.Items) != 1 || len(order.Items[0].DownloadURLs) != 1 {
		t.Fatalf("unexpected order: %+v", order)
	}

	protectedFile, err := integrationRepo.ProductFileForCustomerOrder(
		order.ID, customer.ID, productFile.ID,
	)
	if err != nil || protectedFile.ObjectName != "product-files/lighthouse_aniva/integration.pdf" {
		t.Fatalf("get protected file: file=%+v err=%v", protectedFile, err)
	}
	if _, err := integrationRepo.ProductFileForCustomerOrder(order.ID, customer.ID+1, productFile.ID); err == nil {
		t.Fatal("expected protected file access to fail for another customer")
	} else {
		assertAppError(t, err, models.ErrNotFound, "product_file_not_found")
	}

	orders, err := integrationRepo.Orders()
	if err != nil || !containsOrder(orders, order.ID) {
		t.Fatalf("admin orders missing created order: orders=%+v err=%v", orders, err)
	}
	customerOrders, err := integrationRepo.CustomerOrders(customer.ID)
	if err != nil || !containsOrder(customerOrders, order.ID) {
		t.Fatalf("customer orders missing created order: orders=%+v err=%v", customerOrders, err)
	}
	if err := integrationRepo.DeleteCustomerSession(customerSession.ID); err != nil {
		t.Fatalf("delete customer session: %v", err)
	}
}

func TestCustomerRegistrationAndPasswordResetCodes(t *testing.T) {
	const email = "registration@example.com"
	now := time.Now()
	if err := integrationRepo.SaveCustomerRegistrationCode(
		email, "Registration User", "initial-hash", "registration-code", now.Add(time.Hour),
	); err != nil {
		t.Fatalf("save registration code: %v", err)
	}
	if _, err := integrationRepo.CustomerByRegistrationCode(email, "wrong-code", now); err == nil {
		t.Fatal("expected wrong registration code to fail")
	}
	customer, err := integrationRepo.CustomerByRegistrationCode(email, "registration-code", now)
	if err != nil || customer.Email != email || customer.PasswordHash != "initial-hash" {
		t.Fatalf("verify registration code: customer=%+v err=%v", customer, err)
	}
	if err := integrationRepo.DeleteCustomerRegistrationCode(email); err != nil {
		t.Fatalf("delete registration code: %v", err)
	}

	if err := integrationRepo.SaveCustomerPasswordResetCode(
		email, "reset-code", now.Add(time.Hour),
	); err != nil {
		t.Fatalf("save reset code: %v", err)
	}
	if _, err := integrationRepo.UpdateCustomerPasswordByResetCode(
		email, "wrong-code", "new-hash", now,
	); err == nil {
		t.Fatal("expected wrong password reset code to fail")
	}
	customer, err = integrationRepo.UpdateCustomerPasswordByResetCode(
		email, "reset-code", "new-hash", now,
	)
	if err != nil || customer.PasswordHash != "new-hash" {
		t.Fatalf("verify reset code: customer=%+v err=%v", customer, err)
	}
	if err := integrationRepo.DeleteCustomerPasswordResetCode(email); err != nil {
		t.Fatalf("delete reset code: %v", err)
	}
}

func assertAppError(t *testing.T, err error, kind error, code string) {
	t.Helper()
	if !errors.Is(err, kind) {
		t.Fatalf("expected error kind %v, got %v", kind, err)
	}
	var appErr models.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected error code %q, got %q", code, appErr.Code)
	}
}

func containsGallery(items []models.GalleryItem, id int64, title string, imageURL string) bool {
	for _, item := range items {
		if item.ID == id && item.Title == title && item.Img == imageURL {
			return true
		}
	}
	return false
}

func containsPost(posts []models.BlogPost, id string, title string, imageURL string) bool {
	for _, post := range posts {
		if post.ID == id && post.Title == title && post.Img == imageURL {
			return true
		}
	}
	return false
}

func containsTestimonial(items []models.Testimonial, id int64, name string, imageURL string) bool {
	for _, item := range items {
		if item.ID == id && item.Name == name && item.Img == imageURL {
			return true
		}
	}
	return false
}

func containsOrder(orders []models.Order, id string) bool {
	for _, order := range orders {
		if order.ID == id {
			return true
		}
	}
	return false
}
