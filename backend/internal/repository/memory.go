package repository

import (
	"fmt"
	"sync"
	"time"

	"new-forstitch-site/backend/internal/models"
)

type MemoryRepository struct {
	mu          sync.RWMutex
	adminUsers  []models.AdminUser
	sessions    map[string]models.AdminSession
	categories  []models.Category
	products    []models.Product
	gallery     []models.GalleryItem
	blog        []models.BlogPost
	siteContent models.SiteContent
	orders      []models.OrderRequest
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		sessions: map[string]models.AdminSession{},
		categories: []models.Category{
			{ID: "fauna", Label: "Животный мир"},
			{ID: "people", Label: "Люди"},
			{ID: "still-life", Label: "Натюрморты"},
			{ID: "landscape", Label: "Пейзаж"},
			{ID: "fantasy", Label: "Фэнтези"},
		},
		products: []models.Product{
			{
				ID:     "lighthouse_aniva",
				Title:  "Маяк на мысе Анива",
				Price:  600,
				Cat:    "landscape",
				Sub:    "Море",
				Img:    "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				IsNew:  true,
				Size:   "300 x 220 крестов",
				Colors: "58 цветов DMC",
				Canvas: "Aida 16 / равномерка 32",
			},
			{
				ID:     "oxota_na_miod",
				Title:  "Охота на мед",
				Price:  200,
				Cat:    "fauna",
				Sub:    "Насекомые",
				Img:    "https://forstitch.ru/wp-content/uploads/2021/04/5-300x300.jpg",
				IsNew:  true,
				Size:   "120 x 120 крестов",
				Colors: "32 цвета DMC",
				Canvas: "Aida 14",
			},
			{
				ID:     "dragon_library",
				Title:  "Дракон-читальня",
				Price:  450,
				Cat:    "fantasy",
				Sub:    "Драконы",
				Img:    "https://forstitch.ru/wp-content/uploads/2016/11/8SNwJDfXaw-1030x833.jpg",
				Size:   "240 x 190 крестов",
				Colors: "52 цвета DMC",
				Canvas: "Aida 16",
			},
			{
				ID:     "anemones",
				Title:  "Анемоны",
				Price:  400,
				Cat:    "still-life",
				Sub:    "Цветы",
				Img:    "https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg",
				Size:   "180 x 240 крестов",
				Colors: "46 цветов DMC",
				Canvas: "равномерка 32",
			},
		},
		gallery: []models.GalleryItem{
			{
				Img:   "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				Title: "Маяк на мысе Анива",
				By:    "Команда Forstitch",
			},
			{
				Img:   "https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg",
				Title: "Анемоны",
				By:    "Команда Forstitch",
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
			},
		},
		siteContent: models.SiteContent{
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
	return clone(s.products)
}

func (s *MemoryRepository) Product(id string) (models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, product := range s.products {
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
			s.products[index] = product
			return nil
		}
	}
	return models.NotFound("product_not_found", "product not found")
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

func (s *MemoryRepository) Blog() []models.BlogPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.blog)
}

func (s *MemoryRepository) SiteContent() models.SiteContent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.siteContent
}

func (s *MemoryRepository) CreateOrder(req models.OrderRequest) models.OrderResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders = append(s.orders, req)
	return models.OrderResponse{
		ID:      fmt.Sprintf("order_%d_%d", time.Now().Unix(), len(s.orders)),
		Message: "Заказ создан в тестовом backend. Оплата будет подключена позже.",
	}
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

func clone[T any](items []T) []T {
	out := make([]T, len(items))
	copy(out, items)
	return out
}
