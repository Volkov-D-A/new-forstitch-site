package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"new-forstitch-site/backend/internal/site"
)

var ErrNotFound = errors.New("not found")

type MemoryStore struct {
	mu          sync.RWMutex
	categories  []site.Category
	products    []site.Product
	gallery     []site.GalleryItem
	blog        []site.BlogPost
	siteContent site.SiteContent
	orders      []site.OrderRequest
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		categories: []site.Category{
			{ID: "fauna", Label: "Животный мир"},
			{ID: "people", Label: "Люди"},
			{ID: "still-life", Label: "Натюрморты"},
			{ID: "landscape", Label: "Пейзаж"},
			{ID: "fantasy", Label: "Фэнтези"},
		},
		products: []site.Product{
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
		gallery: []site.GalleryItem{
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
		blog: []site.BlogPost{
			{
				ID:      "new-patterns",
				Title:   "Новые схемы в каталоге",
				Date:    "2026-06-10",
				Tag:     "Новости",
				Img:     "https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg",
				Excerpt: "Первые товары уже отдаются из Go API. Дальше сюда можно подключить админку и базу данных.",
			},
		},
		siteContent: site.SiteContent{
			Author: site.Author{
				Name:  "Екатерина Волкова",
				Photo: "https://forstitch.ru/wp-content/uploads/2016/04/MG_4272-687x1030.jpg",
				P1:    "Авторские схемы для вышивки крестом с вниманием к цвету, деталям и удобству отшива.",
				P2:    "Каждая схема готовится вручную и проверяется перед публикацией.",
				P3:    "Сайт постепенно переезжает на новый backend, чтобы каталогом было удобно управлять.",
				Sign:  "Екатерина",
			},
			HowToBuy: []site.HowToStep{
				{N: "01", T: "Выберите схему", D: "Добавьте понравившуюся PDF-схему в корзину."},
				{N: "02", T: "Оформите заказ", D: "Backend создаст заказ и подготовит ссылку на оплату."},
				{N: "03", T: "Оплатите", D: "После оплаты схема будет отправлена на указанную почту."},
				{N: "04", T: "Вышивайте", D: "Откройте PDF-файл, подготовьте материалы и начинайте отшив."},
			},
			Testimonials: []site.Testimonial{
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

func (s *MemoryStore) Categories() []site.Category {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.categories)
}

func (s *MemoryStore) Products() []site.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.products)
}

func (s *MemoryStore) Product(id string) (site.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, product := range s.products {
		if product.ID == id {
			return product, nil
		}
	}
	return site.Product{}, ErrNotFound
}

func (s *MemoryStore) Gallery() []site.GalleryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.gallery)
}

func (s *MemoryStore) Blog() []site.BlogPost {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.blog)
}

func (s *MemoryStore) SiteContent() site.SiteContent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.siteContent
}

func (s *MemoryStore) CreateOrder(req site.OrderRequest) site.OrderResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders = append(s.orders, req)
	return site.OrderResponse{
		ID:      fmt.Sprintf("order_%d_%d", time.Now().Unix(), len(s.orders)),
		Message: "Заказ создан в тестовом backend. Оплата будет подключена позже.",
	}
}

func clone[T any](items []T) []T {
	out := make([]T, len(items))
	copy(out, items)
	return out
}
