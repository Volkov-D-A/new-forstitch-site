package services

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"new-forstitch-site/backend/internal/models"
	"new-forstitch-site/backend/internal/repository"
)

const maxProductImageSize = 10 << 20

type FileStorage interface {
	PutProductImage(ctx context.Context, productID string, filename string, contentType string, reader io.Reader, size int64) (string, error)
	PutTestimonialImage(ctx context.Context, testimonialID string, filename string, contentType string, reader io.Reader, size int64) (string, error)
	PutBlogImage(ctx context.Context, postID string, filename string, contentType string, reader io.Reader, size int64) (string, error)
	PutGalleryImage(ctx context.Context, itemID string, filename string, contentType string, reader io.Reader, size int64) (string, error)
	Get(ctx context.Context, objectName string) (io.ReadCloser, models.FileObject, error)
}

type Service struct {
	fileBaseURL string
	files       FileStorage
	repo        repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ConfigureFiles(files FileStorage, fileBaseURL string) {
	s.files = files
	s.fileBaseURL = strings.TrimRight(fileBaseURL, "/")
}

func (s *Service) Categories() []models.Category {
	return s.repo.Categories()
}

func (s *Service) CreateCategory(category models.Category) (models.Category, error) {
	category.ID = newID()
	if err := validateCategory(category); err != nil {
		return models.Category{}, err
	}
	if err := s.repo.CreateCategory(category); err != nil {
		return models.Category{}, err
	}
	return category, nil
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

func (s *Service) CreateProduct(product models.Product) (models.Product, error) {
	product.ID = newID()
	if err := validateProduct(product); err != nil {
		return models.Product{}, err
	}
	if err := s.repo.CreateProduct(product); err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (s *Service) UpdateProduct(id string, product models.Product) error {
	product.ID = id
	if err := validateProduct(product); err != nil {
		return err
	}
	return s.repo.UpdateProduct(id, product)
}

func (s *Service) UploadProductImage(ctx context.Context, id string, filename string, contentType string, reader io.Reader, size int64) (models.Product, error) {
	if strings.TrimSpace(id) == "" {
		return models.Product{}, validation("id_required", "id is required")
	}
	if err := s.validateImageUpload(reader, contentType, size); err != nil {
		return models.Product{}, err
	}

	objectName, err := s.files.PutProductImage(ctx, id, filename, contentType, reader, size)
	if err != nil {
		return models.Product{}, err
	}

	imageURL := s.fileBaseURL + "/" + objectName
	if err := s.repo.UpdateProductImage(id, imageURL); err != nil {
		return models.Product{}, err
	}
	return s.repo.Product(id)
}

func (s *Service) validateImageUpload(reader io.Reader, contentType string, size int64) error {
	if s.files == nil || s.fileBaseURL == "" {
		return models.Internal("file_storage_not_configured", "file storage is not configured")
	}
	if reader == nil {
		return models.BadRequest("file_required", "file is required")
	}
	if size <= 0 {
		return models.BadRequest("file_empty", "file must not be empty")
	}
	if size > maxProductImageSize {
		return models.BadRequest("file_too_large", "file must be 10MB or smaller")
	}
	if !strings.HasPrefix(contentType, "image/") {
		return models.BadRequest("file_type_invalid", "file must be an image")
	}
	return nil
}

func (s *Service) File(ctx context.Context, objectName string) (io.ReadCloser, models.FileObject, error) {
	objectName = strings.Trim(strings.TrimSpace(objectName), "/")
	if objectName == "" {
		return nil, models.FileObject{}, models.NotFound("file_not_found", "file not found")
	}
	if s.files == nil {
		return nil, models.FileObject{}, models.NotFound("file_not_found", "file not found")
	}
	reader, info, err := s.files.Get(ctx, objectName)
	if err != nil {
		return nil, models.FileObject{}, models.NotFound("file_not_found", "file not found")
	}
	return reader, info, nil
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

func (s *Service) CreateGalleryItem(item models.GalleryItem) (models.GalleryItem, error) {
	item = normalizeGalleryItem(item)
	if err := validateGalleryItem(item); err != nil {
		return models.GalleryItem{}, err
	}
	return s.repo.CreateGalleryItem(item)
}

func (s *Service) UpdateGalleryItem(id int64, item models.GalleryItem) (models.GalleryItem, error) {
	if id <= 0 {
		return models.GalleryItem{}, validation("id_required", "id is required")
	}
	item = normalizeGalleryItem(item)
	if err := validateGalleryItem(item); err != nil {
		return models.GalleryItem{}, err
	}
	if err := s.repo.UpdateGalleryItem(id, item); err != nil {
		return models.GalleryItem{}, err
	}
	item.ID = id
	return item, nil
}

func (s *Service) DeleteGalleryItem(id int64) error {
	if id <= 0 {
		return validation("id_required", "id is required")
	}
	return s.repo.DeleteGalleryItem(id)
}

func (s *Service) UploadGalleryItemImage(ctx context.Context, id int64, filename string, contentType string, reader io.Reader, size int64) (models.GalleryItem, error) {
	if id <= 0 {
		return models.GalleryItem{}, validation("id_required", "id is required")
	}
	if err := s.validateImageUpload(reader, contentType, size); err != nil {
		return models.GalleryItem{}, err
	}

	objectName, err := s.files.PutGalleryImage(ctx, strconv.FormatInt(id, 10), filename, contentType, reader, size)
	if err != nil {
		return models.GalleryItem{}, err
	}
	imageURL := s.fileBaseURL + "/" + objectName
	if err := s.repo.UpdateGalleryItemImage(id, imageURL); err != nil {
		return models.GalleryItem{}, err
	}
	for _, item := range s.repo.Gallery() {
		if item.ID == id {
			return item, nil
		}
	}
	return models.GalleryItem{}, models.NotFound("gallery_item_not_found", "gallery item not found")
}

func (s *Service) Blog() []models.BlogPost {
	return s.repo.Blog()
}

func (s *Service) CreateBlogPost(post models.BlogPost) (models.BlogPost, error) {
	post.ID = newID()
	post = normalizeBlogPost(post)
	if err := validateBlogPost(post); err != nil {
		return models.BlogPost{}, err
	}
	return s.repo.CreateBlogPost(post)
}

func (s *Service) UpdateBlogPost(id string, post models.BlogPost) (models.BlogPost, error) {
	if strings.TrimSpace(id) == "" {
		return models.BlogPost{}, validation("id_required", "id is required")
	}
	post.ID = id
	post = normalizeBlogPost(post)
	if err := validateBlogPost(post); err != nil {
		return models.BlogPost{}, err
	}
	if err := s.repo.UpdateBlogPost(id, post); err != nil {
		return models.BlogPost{}, err
	}
	return post, nil
}

func (s *Service) DeleteBlogPost(id string) error {
	if strings.TrimSpace(id) == "" {
		return validation("id_required", "id is required")
	}
	return s.repo.DeleteBlogPost(id)
}

func (s *Service) UploadBlogPostImage(ctx context.Context, id string, filename string, contentType string, reader io.Reader, size int64) (models.BlogPost, error) {
	if strings.TrimSpace(id) == "" {
		return models.BlogPost{}, validation("id_required", "id is required")
	}
	if err := s.validateImageUpload(reader, contentType, size); err != nil {
		return models.BlogPost{}, err
	}

	objectName, err := s.files.PutBlogImage(ctx, id, filename, contentType, reader, size)
	if err != nil {
		return models.BlogPost{}, err
	}

	imageURL := s.fileBaseURL + "/" + objectName
	if err := s.repo.UpdateBlogPostImage(id, imageURL); err != nil {
		return models.BlogPost{}, err
	}

	for _, post := range s.repo.Blog() {
		if post.ID == id {
			return post, nil
		}
	}
	return models.BlogPost{}, models.NotFound("blog_post_not_found", "blog post not found")
}

func (s *Service) SiteContent() models.SiteContent {
	return s.repo.SiteContent()
}

func (s *Service) Testimonials() []models.Testimonial {
	return s.repo.Testimonials()
}

func (s *Service) CreateTestimonial(testimonial models.Testimonial) (models.Testimonial, error) {
	testimonial = normalizeTestimonial(testimonial)
	if err := validateTestimonial(testimonial); err != nil {
		return models.Testimonial{}, err
	}
	return s.repo.CreateTestimonial(testimonial)
}

func (s *Service) UpdateTestimonial(id int64, testimonial models.Testimonial) (models.Testimonial, error) {
	if id <= 0 {
		return models.Testimonial{}, validation("id_required", "id is required")
	}
	testimonial = normalizeTestimonial(testimonial)
	if err := validateTestimonial(testimonial); err != nil {
		return models.Testimonial{}, err
	}
	if err := s.repo.UpdateTestimonial(id, testimonial); err != nil {
		return models.Testimonial{}, err
	}
	testimonial.ID = id
	return testimonial, nil
}

func (s *Service) DeleteTestimonial(id int64) error {
	if id <= 0 {
		return validation("id_required", "id is required")
	}
	return s.repo.DeleteTestimonial(id)
}

func (s *Service) UploadTestimonialImage(ctx context.Context, id int64, filename string, contentType string, reader io.Reader, size int64) (models.Testimonial, error) {
	if id <= 0 {
		return models.Testimonial{}, validation("id_required", "id is required")
	}
	if err := s.validateImageUpload(reader, contentType, size); err != nil {
		return models.Testimonial{}, err
	}

	objectName, err := s.files.PutTestimonialImage(ctx, strconv.FormatInt(id, 10), filename, contentType, reader, size)
	if err != nil {
		return models.Testimonial{}, err
	}

	imageURL := s.fileBaseURL + "/" + objectName
	if err := s.repo.UpdateTestimonialImage(id, imageURL); err != nil {
		return models.Testimonial{}, err
	}

	for _, testimonial := range s.repo.Testimonials() {
		if testimonial.ID == id {
			return testimonial, nil
		}
	}
	return models.Testimonial{}, models.NotFound("testimonial_not_found", "testimonial not found")
}

func (s *Service) SiteSettings() models.SiteSettings {
	content := s.repo.SiteContent()
	return models.SiteSettings{FeaturedProductID: content.FeaturedProductID}
}

func (s *Service) UpdateSiteSettings(settings models.SiteSettings) (models.SiteSettings, error) {
	settings.FeaturedProductID = strings.TrimSpace(settings.FeaturedProductID)
	if settings.FeaturedProductID != "" {
		if _, err := s.repo.Product(settings.FeaturedProductID); err != nil {
			return models.SiteSettings{}, err
		}
	}
	if err := s.repo.UpdateSiteSettings(settings); err != nil {
		return models.SiteSettings{}, err
	}
	return settings, nil
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

func normalizeGalleryItem(item models.GalleryItem) models.GalleryItem {
	item.Img = strings.TrimSpace(item.Img)
	item.Title = strings.TrimSpace(item.Title)
	item.By = strings.TrimSpace(item.By)
	return item
}

func validateGalleryItem(item models.GalleryItem) error {
	if item.Title == "" {
		return validation("title_required", "title is required")
	}
	if item.By == "" {
		return validation("author_required", "by is required")
	}
	return nil
}

func normalizeBlogPost(post models.BlogPost) models.BlogPost {
	post.Title = strings.TrimSpace(post.Title)
	post.Date = strings.TrimSpace(post.Date)
	post.Tag = strings.TrimSpace(post.Tag)
	post.Img = strings.TrimSpace(post.Img)
	post.Excerpt = strings.TrimSpace(post.Excerpt)
	post.Content = strings.TrimSpace(post.Content)
	return post
}

func validateBlogPost(post models.BlogPost) error {
	if strings.TrimSpace(post.ID) == "" {
		return validation("id_required", "id is required")
	}
	if post.Title == "" {
		return validation("title_required", "title is required")
	}
	if post.Date == "" {
		return validation("date_required", "date is required")
	}
	if _, err := time.Parse("2006-01-02", post.Date); err != nil {
		return validation("date_invalid", "date must use YYYY-MM-DD format")
	}
	if post.Excerpt == "" {
		return validation("excerpt_required", "excerpt is required")
	}
	if post.Content == "" {
		return validation("content_required", "content is required")
	}
	return nil
}

func normalizeTestimonial(testimonial models.Testimonial) models.Testimonial {
	testimonial.Name = strings.TrimSpace(testimonial.Name)
	testimonial.Role = strings.TrimSpace(testimonial.Role)
	testimonial.Img = strings.TrimSpace(testimonial.Img)
	testimonial.Text = strings.TrimSpace(testimonial.Text)
	return testimonial
}

func validateTestimonial(testimonial models.Testimonial) error {
	if testimonial.Name == "" {
		return validation("name_required", "name is required")
	}
	if testimonial.Text == "" {
		return validation("text_required", "text is required")
	}
	return nil
}

func validation(code string, message string) error {
	return models.Validation(code, message)
}

func newID() string {
	return uuid.NewString()
}
