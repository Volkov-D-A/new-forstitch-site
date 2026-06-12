package repository

import (
	"fmt"
	"sync"
	"time"

	"new-forstitch-site/backend/internal/models"
)

type MemoryRepository struct {
	mu                sync.RWMutex
	adminUsers        []models.AdminUser
	sessions          map[string]models.AdminSession
	customers         []models.CustomerUser
	customerSessions  map[string]models.CustomerSession
	registrationCodes map[string]memoryRegistrationCode
	resetCodes        map[string]memoryPasswordResetCode
	categories        []models.Category
	products          []models.Product
	gallery           []models.GalleryItem
	blog              []models.BlogPost
	siteContent       models.SiteContent
	orders            []models.Order
}

type memoryRegistrationCode struct {
	Name         string
	PasswordHash string
	CodeHash     string
	ExpiresAt    time.Time
}

type memoryPasswordResetCode struct {
	CodeHash  string
	ExpiresAt time.Time
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		sessions:          map[string]models.AdminSession{},
		customerSessions:  map[string]models.CustomerSession{},
		registrationCodes: map[string]memoryRegistrationCode{},
		resetCodes:        map[string]memoryPasswordResetCode{},
		categories: []models.Category{
			{ID: "fauna", Label: "Животный мир"},
			{ID: "people", Label: "Люди"},
			{ID: "still-life", Label: "Натюрморты"},
			{ID: "landscape", Label: "Пейзаж"},
			{ID: "fantasy", Label: "Фэнтези"},
		},
		products: []models.Product{
			{
				ID:          "lighthouse_aniva",
				Title:       "Маяк на мысе Анива",
				Price:       600,
				Cat:         "landscape",
				Img:         "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				IsNew:       true,
				Size:        "300 x 220 крестов",
				Colors:      "58 цветов DMC",
				Description: "Пейзажная схема с мягкими переходами и морским светом.",
			},
			{
				ID:          "oxota_na_miod",
				Title:       "Охота на мед",
				Price:       200,
				Cat:         "fauna",
				Img:         "https://forstitch.ru/wp-content/uploads/2021/04/5-300x300.jpg",
				IsNew:       true,
				Size:        "120 x 120 крестов",
				Colors:      "32 цвета DMC",
				Description: "Небольшая схема для быстрого уютного проекта.",
			},
			{
				ID:          "dragon_library",
				Title:       "Дракон-читальня",
				Price:       450,
				Cat:         "fantasy",
				Img:         "https://forstitch.ru/wp-content/uploads/2016/11/8SNwJDfXaw-1030x833.jpg",
				Size:        "240 x 190 крестов",
				Colors:      "52 цвета DMC",
				Description: "Фэнтезийный сюжет с детализированной книжной полкой.",
			},
			{
				ID:          "anemones",
				Title:       "Анемоны",
				Price:       400,
				Cat:         "still-life",
				Img:         "https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg",
				Size:        "180 x 240 крестов",
				Colors:      "46 цветов DMC",
				Description: "Цветочный натюрморт с выразительными оттенками.",
			},
		},
		gallery: []models.GalleryItem{
			{
				ID:          1,
				Img:         "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				Title:       "Маяк на мысе Анива",
				Description: "Готовый отшив схемы с маяком на скалистом берегу.",
			},
			{
				ID:          2,
				Img:         "https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg",
				Title:       "Анемоны",
				Description: "Цветочная вышивка с яркими анемонами.",
			},
		},
		blog: []models.BlogPost{
			{
				ID:      "new-patterns",
				Title:   "Новые схемы в каталоге",
				Date:    "2026-06-10",
				Tag:     "Новости",
				Img:     "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				Excerpt: "Первые товары уже отдаются из Go API. Дальше сюда можно подключить админку и базу данных.",
				Content: "Первые товары уже отдаются из Go API. Дальше сюда можно подключить админку и базу данных.",
			},
		},
		siteContent: models.SiteContent{
			FeaturedProductID: "lighthouse_aniva",
			Author: models.Author{
				Name:  "Екатерина Волкова",
				Photo: "https://forstitch.ru/wp-content/uploads/2016/04/MG_4272-687x1030.jpg",
				P1:    "Авторские схемы для вышивки крестом с вниманием к цвету, деталям и удобству отшива.",
				P2:    "Каждая схема готовится вручную и проверяется перед публикацией.",
				P3:    "Сайт постепенно переезжает на новый backend, чтобы каталогом было удобно управлять.",
				Sign:  "Екатерина",
			},
			HowToBuy: []models.HowToStep{
				{N: "01", T: "Выберите схему", D: "Добавьте понравившуюся PDF-схему в корзину."},
				{N: "02", T: "Оформите заказ", D: "Backend создаст заказ и подготовит ссылку на оплату."},
				{N: "03", T: "Оплатите", D: "После оплаты схема будет отправлена на указанную почту."},
				{N: "04", T: "Вышивайте", D: "Откройте PDF-файл, подготовьте материалы и начинайте отшив."},
			},
			Testimonials: []models.Testimonial{
				{
					ID:   1,
					Name: "Мария",
					Role: "Вышивальщица",
					Img:  "https://forstitch.ru/wp-content/uploads/2021/04/5-300x300.jpg",
					Text: "Плавные переходы и понятная схема, приятно вышивать.",
				},
			},
		},
	}
}

func (s *MemoryRepository) Categories() []models.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.categories)
}

func (s *MemoryRepository) CreateCategory(category models.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.categories {
		if item.ID == category.ID {
			return models.Conflict("category_exists", "category already exists")
		}
	}
	s.categories = append(s.categories, category)
	return nil
}

func (s *MemoryRepository) UpdateCategory(id string, category models.Category) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.categories {
		if item.ID == id {
			s.categories[index] = category
			return nil
		}
	}
	return models.NotFound("category_not_found", "category not found")
}

func (s *MemoryRepository) DeleteCategory(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, product := range s.products {
		if product.Cat == id {
			return models.Conflict("category_has_products", "category has products")
		}
	}
	for index, item := range s.categories {
		if item.ID == id {
			s.categories = append(s.categories[:index], s.categories[index+1:]...)
			return nil
		}
	}
	return models.NotFound("category_not_found", "category not found")
}

func (s *MemoryRepository) Products() []models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return markLatestProducts(clone(s.products))
}

func (s *MemoryRepository) Product(id string) (models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, product := range markLatestProducts(clone(s.products)) {
		if product.ID == id {
			return product, nil
		}
	}
	return models.Product{}, models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) CreateProduct(product models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.products {
		if item.ID == product.ID {
			return models.Conflict("product_exists", "product already exists")
		}
	}
	s.products = append(s.products, product)
	return nil
}

func (s *MemoryRepository) UpdateProduct(id string, product models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.products {
		if item.ID == id {
			product.Images = item.Images
			product.Files = item.Files
			s.products[index] = product
			return nil
		}
	}
	return models.NotFound("product_not_found", "product not found")
}

func markLatestProducts(products []models.Product) []models.Product {
	for index := range products {
		products[index].IsNew = index >= len(products)-4
	}
	return products
}

func (s *MemoryRepository) UpdateProductImage(id string, imageURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.products {
		if item.ID == id {
			s.products[index].Img = imageURL
			return nil
		}
	}
	return models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) AddProductImage(productID string, imageURL string) (models.ProductImage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, product := range s.products {
		if product.ID == productID {
			image := models.ProductImage{ID: int64(len(product.Images) + 1), URL: imageURL}
			s.products[index].Images = append(s.products[index].Images, image)
			return image, nil
		}
	}
	return models.ProductImage{}, models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) DeleteProductImage(productID string, imageID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for productIndex, product := range s.products {
		if product.ID != productID {
			continue
		}
		for imageIndex, image := range product.Images {
			if image.ID == imageID {
				s.products[productIndex].Images = append(product.Images[:imageIndex], product.Images[imageIndex+1:]...)
				return nil
			}
		}
		return models.NotFound("product_image_not_found", "product image not found")
	}
	return models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) AddProductFile(productID string, name string, objectName string) (models.ProductFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, product := range s.products {
		if product.ID == productID {
			file := models.ProductFile{
				ID:         int64(len(product.Files) + 1),
				Name:       name,
				ObjectName: objectName,
			}
			s.products[index].Files = append(s.products[index].Files, file)
			return file, nil
		}
	}
	return models.ProductFile{}, models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) DeleteProductFile(productID string, fileID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for productIndex, product := range s.products {
		if product.ID != productID {
			continue
		}
		for fileIndex, file := range product.Files {
			if file.ID == fileID {
				s.products[productIndex].Files = append(product.Files[:fileIndex], product.Files[fileIndex+1:]...)
				return nil
			}
		}
		return models.NotFound("product_file_not_found", "product file not found")
	}
	return models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) ProductFileForCustomerOrder(orderID string, _ int64, fileID int64) (models.ProductFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, order := range s.orders {
		if order.ID != orderID || (order.Status != "paid" && order.Status != "fulfilled") {
			continue
		}
		for _, item := range order.Items {
			for _, product := range s.products {
				if product.ID != item.ProductID {
					continue
				}
				for _, file := range product.Files {
					if file.ID == fileID {
						return file, nil
					}
				}
			}
		}
	}
	return models.ProductFile{}, models.NotFound("product_file_not_found", "product file not found")
}

func (s *MemoryRepository) DeleteProduct(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.products {
		if item.ID == id {
			s.products = append(s.products[:index], s.products[index+1:]...)
			return nil
		}
	}
	return models.NotFound("product_not_found", "product not found")
}

func (s *MemoryRepository) Gallery() []models.GalleryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.gallery)
}

func (s *MemoryRepository) CreateGalleryItem(item models.GalleryItem) (models.GalleryItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var maxID int64
	for _, existing := range s.gallery {
		if existing.ID > maxID {
			maxID = existing.ID
		}
	}
	item.ID = maxID + 1
	s.gallery = append(s.gallery, item)
	return item, nil
}

func (s *MemoryRepository) UpdateGalleryItem(id int64, item models.GalleryItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.gallery {
		if existing.ID == id {
			item.ID = id
			s.gallery[index] = item
			return nil
		}
	}
	return models.NotFound("gallery_item_not_found", "gallery item not found")
}

func (s *MemoryRepository) UpdateGalleryItemImage(id int64, imageURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.gallery {
		if item.ID == id {
			s.gallery[index].Img = imageURL
			return nil
		}
	}
	return models.NotFound("gallery_item_not_found", "gallery item not found")
}

func (s *MemoryRepository) DeleteGalleryItem(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.gallery {
		if item.ID == id {
			s.gallery = append(s.gallery[:index], s.gallery[index+1:]...)
			return nil
		}
	}
	return models.NotFound("gallery_item_not_found", "gallery item not found")
}

func (s *MemoryRepository) Blog() []models.BlogPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.blog)
}

func (s *MemoryRepository) CreateBlogPost(post models.BlogPost) (models.BlogPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.blog {
		if item.ID == post.ID {
			return models.BlogPost{}, models.Conflict("blog_post_exists", "blog post already exists")
		}
	}
	s.blog = append(s.blog, post)
	return post, nil
}

func (s *MemoryRepository) UpdateBlogPost(id string, post models.BlogPost) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.blog {
		if item.ID == id {
			post.ID = id
			s.blog[index] = post
			return nil
		}
	}
	return models.NotFound("blog_post_not_found", "blog post not found")
}

func (s *MemoryRepository) UpdateBlogPostImage(id string, imageURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.blog {
		if item.ID == id {
			s.blog[index].Img = imageURL
			return nil
		}
	}
	return models.NotFound("blog_post_not_found", "blog post not found")
}

func (s *MemoryRepository) DeleteBlogPost(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.blog {
		if item.ID == id {
			s.blog = append(s.blog[:index], s.blog[index+1:]...)
			return nil
		}
	}
	return models.NotFound("blog_post_not_found", "blog post not found")
}

func (s *MemoryRepository) SiteContent() models.SiteContent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.siteContent
}

func (s *MemoryRepository) Testimonials() []models.Testimonial {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.siteContent.Testimonials)
}

func (s *MemoryRepository) CreateTestimonial(testimonial models.Testimonial) (models.Testimonial, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var maxID int64
	for _, item := range s.siteContent.Testimonials {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	testimonial.ID = maxID + 1
	s.siteContent.Testimonials = append(s.siteContent.Testimonials, testimonial)
	return testimonial, nil
}

func (s *MemoryRepository) UpdateTestimonial(id int64, testimonial models.Testimonial) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.siteContent.Testimonials {
		if item.ID == id {
			testimonial.ID = id
			s.siteContent.Testimonials[index] = testimonial
			return nil
		}
	}
	return models.NotFound("testimonial_not_found", "testimonial not found")
}

func (s *MemoryRepository) UpdateTestimonialImage(id int64, imageURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.siteContent.Testimonials {
		if item.ID == id {
			s.siteContent.Testimonials[index].Img = imageURL
			return nil
		}
	}
	return models.NotFound("testimonial_not_found", "testimonial not found")
}

func (s *MemoryRepository) DeleteTestimonial(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.siteContent.Testimonials {
		if item.ID == id {
			s.siteContent.Testimonials = append(s.siteContent.Testimonials[:index], s.siteContent.Testimonials[index+1:]...)
			return nil
		}
	}
	return models.NotFound("testimonial_not_found", "testimonial not found")
}

func (s *MemoryRepository) UpdateSiteSettings(settings models.SiteSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.siteContent.FeaturedProductID = settings.FeaturedProductID
	return nil
}

func (s *MemoryRepository) CreateOrder(req models.OrderRequest, customer models.CustomerUser) (models.OrderResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orderID := fmt.Sprintf("order_%d_%d", time.Now().Unix(), len(s.orders)+1)
	order := models.Order{
		ID:            orderID,
		Status:        "paid",
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		Message:       "Заказ оформлен и считается оплаченным.",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	for _, cartItem := range req.Items {
		item := models.OrderItem{
			ProductID:   cartItem.ProductID,
			ProductName: cartItem.ProductID,
			Quantity:    cartItem.Quantity,
			Price:       s.productPrice(cartItem.ProductID),
		}
		for _, product := range s.products {
			if product.ID != cartItem.ProductID {
				continue
			}
			item.ProductName = product.Title
			for _, file := range product.Files {
				item.DownloadURLs = append(item.DownloadURLs, models.DownloadFile{
					ID:   file.ID,
					Name: file.Name,
					URL:  fmt.Sprintf("/api/customer/orders/%s/files/%d", orderID, file.ID),
				})
			}
		}
		order.Items = append(order.Items, item)
	}
	s.orders = append(s.orders, order)
	return models.OrderResponse{ID: orderID, Status: order.Status, Message: order.Message}, nil
}

func (s *MemoryRepository) Orders() ([]models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Order(nil), s.orders...), nil
}

func (s *MemoryRepository) CustomerOrders(_ int64) ([]models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]models.Order(nil), s.orders...), nil
}

func (s *MemoryRepository) OrderForCustomer(orderID string, _ int64) (models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, order := range s.orders {
		if order.ID == orderID {
			return order, nil
		}
	}
	return models.Order{}, models.NotFound("order_not_found", "order not found")
}

func (s *MemoryRepository) productPrice(productID string) int {
	for _, product := range s.products {
		if product.ID == productID {
			return product.Price
		}
	}
	return 0
}

func (s *MemoryRepository) AdminUserByUsername(username string) (models.AdminUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.adminUsers {
		if user.Username == username {
			return user, nil
		}
	}
	return models.AdminUser{}, models.NotFound("admin_user_not_found", "admin user not found")
}

func (s *MemoryRepository) EnsureAdminUser(username string, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, user := range s.adminUsers {
		if user.Username == username {
			s.adminUsers[index].PasswordHash = passwordHash
			return nil
		}
	}
	s.adminUsers = append(s.adminUsers, models.AdminUser{
		ID:           int64(len(s.adminUsers) + 1),
		Username:     username,
		PasswordHash: passwordHash,
	})
	return nil
}

func (s *MemoryRepository) CreateAdminSession(session models.AdminSession, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryRepository) AdminSession(sessionID string, _ time.Time) (models.AdminSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return models.AdminSession{}, models.Unauthorized("session_invalid", "admin session is invalid")
	}
	return session, nil
}

func (s *MemoryRepository) DeleteAdminSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

func (s *MemoryRepository) DeleteExpiredAdminSessions(_ time.Time) error {
	return nil
}

func (s *MemoryRepository) CustomerByEmail(email string) (models.CustomerUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.customers {
		if user.Email == email {
			return user, nil
		}
	}
	return models.CustomerUser{}, models.NotFound("customer_not_found", "customer not found")
}

func (s *MemoryRepository) EnsureCustomer(email string, name string, passwordHash string) (models.CustomerUser, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, user := range s.customers {
		if user.Email == email {
			if name != "" {
				s.customers[index].Name = name
				user.Name = name
			}
			return user, false, nil
		}
	}
	user := models.CustomerUser{
		ID:           int64(len(s.customers) + 1),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
	}
	s.customers = append(s.customers, user)
	return user, true, nil
}

func (s *MemoryRepository) SaveCustomerRegistrationCode(email string, name string, passwordHash string, codeHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.registrationCodes[email] = memoryRegistrationCode{
		Name:         name,
		PasswordHash: passwordHash,
		CodeHash:     codeHash,
		ExpiresAt:    expiresAt,
	}
	return nil
}

func (s *MemoryRepository) CustomerByRegistrationCode(email string, codeHash string, now time.Time) (models.CustomerUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, ok := s.registrationCodes[email]
	if !ok || code.CodeHash != codeHash || !code.ExpiresAt.After(now) {
		return models.CustomerUser{}, models.NotFound("registration_code_not_found", "registration code not found")
	}
	delete(s.registrationCodes, email)
	for index, user := range s.customers {
		if user.Email == email {
			if code.Name != "" {
				s.customers[index].Name = code.Name
				user.Name = code.Name
			}
			return user, nil
		}
	}
	user := models.CustomerUser{
		ID:           int64(len(s.customers) + 1),
		Email:        email,
		Name:         code.Name,
		PasswordHash: code.PasswordHash,
	}
	s.customers = append(s.customers, user)
	return user, nil
}

func (s *MemoryRepository) DeleteCustomerRegistrationCode(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.registrationCodes, email)
	return nil
}

func (s *MemoryRepository) SaveCustomerPasswordResetCode(email string, codeHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resetCodes[email] = memoryPasswordResetCode{CodeHash: codeHash, ExpiresAt: expiresAt}
	return nil
}

func (s *MemoryRepository) UpdateCustomerPasswordByResetCode(email string, codeHash string, passwordHash string, now time.Time) (models.CustomerUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, ok := s.resetCodes[email]
	if !ok || code.CodeHash != codeHash || !code.ExpiresAt.After(now) {
		return models.CustomerUser{}, models.NotFound("password_reset_code_not_found", "password reset code not found")
	}
	delete(s.resetCodes, email)
	for index, user := range s.customers {
		if user.Email == email {
			s.customers[index].PasswordHash = passwordHash
			user.PasswordHash = passwordHash
			return user, nil
		}
	}
	return models.CustomerUser{}, models.NotFound("customer_not_found", "customer not found")
}

func (s *MemoryRepository) DeleteCustomerPasswordResetCode(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.resetCodes, email)
	return nil
}

func (s *MemoryRepository) CreateCustomerSession(session models.CustomerSession, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.customerSessions[session.ID] = session
	return nil
}

func (s *MemoryRepository) CustomerSession(sessionID string, _ time.Time) (models.CustomerSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.customerSessions[sessionID]
	if !ok {
		return models.CustomerSession{}, models.Unauthorized("session_invalid", "customer session is invalid")
	}
	return session, nil
}

func (s *MemoryRepository) DeleteCustomerSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.customerSessions, sessionID)
	return nil
}

func (s *MemoryRepository) DeleteExpiredCustomerSessions(_ time.Time) error {
	return nil
}

func clone[T any](items []T) []T {
	out := make([]T, len(items))
	copy(out, items)
	return out
}
