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
	DeleteProduct(id string) error
}

type ContentRepository interface {
	Gallery() []models.GalleryItem
	Blog() []models.BlogPost
	SiteContent() models.SiteContent
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
