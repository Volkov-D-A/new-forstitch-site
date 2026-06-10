package services

import (
	"strings"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Categories() []models.Category {
	return s.repo.Categories()
}

func (s *Service) CreateCategory(category models.Category) error {
	if err := validateCategory(category); err != nil {
		return err
	}
	return s.repo.CreateCategory(category)
}

func (s *Service) UpdateCategory(id string, category models.Category) error {
	category.ID = id
	if err := validateCategory(category); err != nil {
		return err
	}
	return s.repo.UpdateCategory(id, category)
}

func (s *Service) DeleteCategory(id string) error {
	if strings.TrimSpace(id) == "" {
		return validation("id_required", "id is required")
	}
	return s.repo.DeleteCategory(id)
}

func (s *Service) Products() []models.Product {
	return s.repo.Products()
}

func (s *Service) Product(id string) (models.Product, error) {
	if strings.TrimSpace(id) == "" {
		return models.Product{}, validation("id_required", "id is required")
	}
	return s.repo.Product(id)
}

func (s *Service) CreateProduct(product models.Product) error {
	if err := validateProduct(product); err != nil {
		return err
	}
	return s.repo.CreateProduct(product)
}

func (s *Service) UpdateProduct(id string, product models.Product) error {
	product.ID = id
	if err := validateProduct(product); err != nil {
		return err
	}
	return s.repo.UpdateProduct(id, product)
}

func (s *Service) DeleteProduct(id string) error {
	if strings.TrimSpace(id) == "" {
		return validation("id_required", "id is required")
	}
	return s.repo.DeleteProduct(id)
}

func (s *Service) Gallery() []models.GalleryItem {
	return s.repo.Gallery()
}

func (s *Service) Blog() []models.BlogPost {
	return s.repo.Blog()
}

func (s *Service) SiteContent() models.SiteContent {
	return s.repo.SiteContent()
}

func (s *Service) CreateOrder(req models.OrderRequest) (models.OrderResponse, error) {
	if err := validateOrder(req); err != nil {
		return models.OrderResponse{}, err
	}
	return s.repo.CreateOrder(req), nil
}

func validateOrder(req models.OrderRequest) error {
	if len(req.Items) == 0 {
		return validation("order_empty", "order must contain at least one item")
	}

	for _, item := range req.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return validation("product_id_required", "productId is required")
		}
		if item.Quantity < 1 {
			return validation("quantity_invalid", "quantity must be greater than zero")
		}
	}

	return nil
}

func validateCategory(category models.Category) error {
	if strings.TrimSpace(category.ID) == "" {
		return validation("id_required", "id is required")
	}
	if strings.TrimSpace(category.Label) == "" {
		return validation("label_required", "label is required")
	}
	return nil
}

func validateProduct(product models.Product) error {
	if strings.TrimSpace(product.ID) == "" {
		return validation("id_required", "id is required")
	}
	if strings.TrimSpace(product.Title) == "" {
		return validation("title_required", "title is required")
	}
	if strings.TrimSpace(product.Cat) == "" {
		return validation("category_required", "cat is required")
	}
	if product.Price < 0 {
		return validation("price_invalid", "price must be greater than or equal to zero")
	}
	return nil
}

func validation(code string, message string) error {
	return models.Validation(code, message)
}
