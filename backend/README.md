# Forstitch Backend

Go API для нового фронтенда Forstitch.

## Запуск

```bash
cd backend
go run ./cmd/api
```

По умолчанию сервер слушает `:3000`.
При запуске API применяет SQL-миграции и читает данные из PostgreSQL.

```bash
HTTP_ADDR=:3001 go run ./cmd/api
```

По умолчанию используется локальная БД:

```bash
postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable
```

Можно переопределить:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable go run ./cmd/api
```

Admin-пользователь создается/обновляется при старте API. Локальные значения по умолчанию:

```bash
ADMIN_USERNAME=admin
ADMIN_PASSWORD=dev-admin-password
```

## Миграции

```bash
go run ./cmd/migrate
```

SQL-файлы лежат в `internal/db/migrations`.

## Архитектура

Backend разложен по слоям:

- `internal/models` — доменные модели и JSON-контракты.
- `internal/repository` — интерфейсы хранилища и реализации PostgreSQL/in-memory.
- `internal/services` — бизнес-логика, валидация и сценарии приложения.
- `internal/api` — HTTP transport: роутинг, auth middleware, JSON request/response.
- `internal/db` — подключение к БД и миграции.
- `cmd/api` — сборка зависимостей и запуск HTTP API.
- `cmd/migrate` — отдельный запуск миграций.

Новый код стоит добавлять по этому направлению: сначала модель/репозиторий, затем service-метод, затем HTTP handler.

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
POST /api/auth/login
GET  /api/auth/session
POST /api/auth/logout

GET    /api/admin/categories
POST   /api/admin/categories
PUT    /api/admin/categories/{categoryID}
DELETE /api/admin/categories/{categoryID}
GET    /api/admin/products
POST   /api/admin/products
PUT    /api/admin/products/{productID}
DELETE /api/admin/products/{productID}
```

Auth устроена через HttpOnly cookie `forstitch_admin_session`. Для `POST`, `PUT`, `DELETE` admin endpoints нужно передавать CSRF-токен из ответа login/session:

```http
X-CSRF-Token: <csrfToken>
```

Пример входа:

```bash
curl -i -X POST http://localhost:3000/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"dev-admin-password"}'
```

Пример admin-запроса:

```bash
curl -X POST http://localhost:3000/api/admin/categories \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'Content-Type: application/json' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -d '{"id":"flowers","label":"Цветы"}'
```

## Ошибки

Ошибки для фронтенда возвращаются в едином формате:

```json
{
  "error": {
    "code": "product_not_found",
    "message": "product not found"
  }
}
```

Доменные ошибки определяются в `internal/models/errors.go`, HTTP-маппинг находится в `internal/api/router.go`.

## Подключение фронтенда

Фронтенд захардкожен на `http://localhost:3000/api`, поэтому backend нужно запускать на порту `3000`.

## Проверка

```bash
go test ./...
go run ./cmd/migrate
go run ./cmd/api
```
