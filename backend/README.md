# Forstitch Backend

Минимальный Go API для нового фронтенда Forstitch.

## Запуск

```bash
cd backend
go run ./cmd/api
```

По умолчанию сервер слушает `:3000`.

```bash
HTTP_ADDR=:3001 go run ./cmd/api
```

## Endpoints

```http
GET  /healthz
GET  /api/categories
GET  /api/products
GET  /api/products/{productID}
GET  /api/gallery
GET  /api/blog
GET  /api/site-content
POST /api/orders
```

## Подключение фронтенда

Фронтенд захардкожен на `http://localhost:3000/api`, поэтому backend нужно запускать на порту `3000`.

Сейчас backend использует in-memory seed-данные; следующий шаг — заменить `internal/store` на постоянное хранилище.

## Проверка

```bash
go test ./...
go run ./cmd/api
```
