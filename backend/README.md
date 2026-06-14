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
ADMIN_USERNAME=dimas
ADMIN_PASSWORD=dimas
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
- `internal/storage` — файловое хранилище MinIO.
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
GET  /api/files/{objectName}
GET  /api/gallery
GET  /api/blog
GET  /api/site-content
POST /api/orders
POST /api/customer/register/start
POST /api/customer/register/verify
POST /api/customer/password-reset/start
POST /api/customer/password-reset/verify
POST /api/customer/login
GET  /api/customer/session
POST /api/customer/logout
GET  /api/customer/orders
GET  /api/customer/orders/{orderID}
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
POST   /api/admin/products/{productID}/image
POST   /api/admin/products/{productID}/images
DELETE /api/admin/products/{productID}/images/{imageID}
POST   /api/admin/products/{productID}/files
DELETE /api/admin/products/{productID}/files/{fileID}
DELETE /api/admin/products/{productID}
GET    /api/admin/blog
POST   /api/admin/blog
PUT    /api/admin/blog/{postID}
POST   /api/admin/blog/{postID}/image
POST   /api/admin/blog/images
DELETE /api/admin/blog/{postID}
GET    /api/admin/gallery
POST   /api/admin/gallery
PUT    /api/admin/gallery/{galleryItemID}
POST   /api/admin/gallery/{galleryItemID}/image
DELETE /api/admin/gallery/{galleryItemID}
GET    /api/admin/site-settings
PUT    /api/admin/site-settings
GET    /api/admin/orders
GET    /api/admin/testimonials
POST   /api/admin/testimonials
PUT    /api/admin/testimonials/{testimonialID}
POST   /api/admin/testimonials/{testimonialID}/image
DELETE /api/admin/testimonials/{testimonialID}
```

Auth устроена через HttpOnly cookie `forstitch_admin_session`. Для `POST`, `PUT`, `DELETE` admin endpoints нужно передавать CSRF-токен из ответа login/session:

```http
X-CSRF-Token: <csrfToken>
```

Пример входа:

```bash
curl -i -X POST http://localhost:3000/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"dimas","password":"dimas"}'
```

Пример admin-запроса:

```bash
curl -X POST http://localhost:3000/api/admin/categories \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'Content-Type: application/json' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -d '{"label":"Цветы"}'
```

При создании категорий и товаров `id` генерируется backend-ом в формате UUID. В `PUT`/`DELETE` используется уже существующий `id` из URL.
Поле товара `isNew` вычисляется backend-ом автоматически: новинками считаются последние 4 добавленных опубликованных товара.

Оформление заказа доступно только покупателю с cookie `forstitch_customer_session`.
Для регистрации покупатель вводит email, имя и пароль, получает код на почту через
`/api/customer/register/start`, затем подтверждает код через `/api/customer/register/verify`.
Восстановление пароля устроено так же: `/api/customer/password-reset/start` отправляет код,
`/api/customer/password-reset/verify` принимает код и новый пароль.

Временный платежный режим: заказ авторизованного покупателя сразу считается оплаченным.
Ссылки на скачивание доступны в личном кабинете; письмом они не отправляются.

```bash
curl http://localhost:3000/api/admin/orders \
  --cookie 'forstitch_admin_session=<session>'
```

Настройки главной страницы:

```json
{ "featuredProductId": "lighthouse_aniva" }
```

Payload для отзывов:

```json
{
  "name": "Анна",
  "role": "Вышивальщица",
  "img": "https://example.com/avatar.jpg",
  "text": "Очень понятная схема."
}
```

Фото отзыва загружается отдельным multipart-запросом:

```bash
curl -X POST http://localhost:3000/api/admin/testimonials/1/image \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -F 'file=@/path/to/avatar.jpg'
```

Изображение товара загружается отдельным multipart-запросом после создания/обновления товара:

```bash
curl -X POST http://localhost:3000/api/admin/products/lighthouse_aniva/image \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -F 'file=@/path/to/image.jpg'
```

Backend сохраняет файл в MinIO и обновляет поле товара `img` публичным URL вида `http://localhost:3000/api/files/products/...`.

Дополнительные изображения товара загружаются отдельными multipart-запросами:

```bash
curl -X POST http://localhost:3000/api/admin/products/lighthouse_aniva/images \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -F 'file=@/path/to/detail.jpg'
```

Цифровые файлы товара загружаются на `/api/admin/products/{productID}/files`. После оплаты покупатель получает отдельную защищенную ссылку на каждый файл в личном кабинете. Скачивание доступно только владельцу оплаченного заказа.

Payload для записей блога:

```json
{
  "title": "Процесс вышивки",
  "date": "2026-06-11",
  "tag": "Блог",
  "img": "",
  "excerpt": "Короткая заметка о процессе.",
  "content": "Полный текст записи."
}
```

Обложка записи блога загружается отдельным multipart-запросом:

```bash
curl -X POST http://localhost:3000/api/admin/blog/<postID>/image \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -F 'file=@/path/to/cover.jpg'
```

Payload для элементов галереи:

```json
{
  "title": "Отшив маяка",
  "description": "Готовый отшив схемы с маяком.",
  "img": ""
}
```

Изображение галереи загружается отдельным multipart-запросом:

```bash
curl -X POST http://localhost:3000/api/admin/gallery/1/image \
  --cookie 'forstitch_admin_session=<session>' \
  -H 'X-CSRF-Token: <csrfToken>' \
  -F 'file=@/path/to/gallery.jpg'
```

## Файловое хранилище

Локально используется MinIO:

```bash
make services-start
```

Переменные окружения:

```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=forstitch
MINIO_SECRET_KEY=forstitch-secret
MINIO_BUCKET=forstitch
MINIO_USE_SSL=false
FILE_BASE_URL=http://localhost:3000/api/files
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

В режиме разработки Vite проксирует относительные запросы `/api` на backend по адресу `http://localhost:3000`.

## Проверка

```bash
go test ./...
go run ./cmd/migrate
go run ./cmd/api
```
