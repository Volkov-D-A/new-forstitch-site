package repository

import (
	"time"

	"new-forstitch-site/backend/internal/models"
)

type CatalogRepository interface {
	Categories() []models.Category
	CreateCategory(category models.Category) error
	UpdateCategory(id string, category models.Category) error
	DeleteCategory(id string) error
	Products() []models.Product
	Product(id string) (models.Product, error)
	CreateProduct(product models.Product) error
	UpdateProduct(id string, product models.Product) error
	UpdateProductImage(id string, imageURL string) error
	DeleteProduct(id string) error
}

type ContentRepository interface {
	Gallery() []models.GalleryItem
	CreateGalleryItem(item models.GalleryItem) (models.GalleryItem, error)
	UpdateGalleryItem(id int64, item models.GalleryItem) error
	UpdateGalleryItemImage(id int64, imageURL string) error
	DeleteGalleryItem(id int64) error
	Blog() []models.BlogPost
	CreateBlogPost(post models.BlogPost) (models.BlogPost, error)
	UpdateBlogPost(id string, post models.BlogPost) error
	UpdateBlogPostImage(id string, imageURL string) error
	DeleteBlogPost(id string) error
	Testimonials() []models.Testimonial
	CreateTestimonial(testimonial models.Testimonial) (models.Testimonial, error)
	UpdateTestimonial(id int64, testimonial models.Testimonial) error
	UpdateTestimonialImage(id int64, imageURL string) error
	DeleteTestimonial(id int64) error
	SiteContent() models.SiteContent
	UpdateSiteSettings(settings models.SiteSettings) error
}

type OrderRepository interface {
	CreateOrder(req models.OrderRequest) models.OrderResponse
}

type Repository interface {
	CatalogRepository
	ContentRepository
	OrderRepository
	AuthRepository
}

type AuthRepository interface {
	AdminUserByUsername(username string) (models.AdminUser, error)
	EnsureAdminUser(username string, passwordHash string) error
	CreateAdminSession(session models.AdminSession, expiresAt time.Time) error
	AdminSession(sessionID string, now time.Time) (models.AdminSession, error)
	DeleteAdminSession(sessionID string) error
	DeleteExpiredAdminSessions(now time.Time) error
}
