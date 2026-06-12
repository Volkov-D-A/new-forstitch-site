# new-forstitch-site

## Frontend

```bash
cd frontend
npm install
npm run dev
```

Фронтенд обращается к API по адресу `http://localhost:3000/api`.
Интерфейс администрирования доступен на `http://localhost:5173/admin`.

Контракт бэкенда описан в `frontend/docs/backend-integration.md`.

## Backend

```bash
cd backend
go run ./cmd/api
```

По умолчанию API доступен на `http://localhost:3000/api`.
Backend подключается к PostgreSQL через `DATABASE_URL`; если переменная не задана, используется локальная строка подключения ниже.
Admin-доступ создается при старте backend из `ADMIN_USERNAME`/`ADMIN_PASSWORD`; локальные значения по умолчанию — `dimas` / `dimas`.
Для cookie-auth backend должен отдавать конкретный CORS origin, не `*`. По умолчанию разрешены `http://localhost:5173` и `http://127.0.0.1:5173`; дополнительные адреса задаются через `CORS_ALLOWED_ORIGINS`.

## Database

```bash
docker compose up -d postgres
```

Или через `Makefile`:

```bash
make db-start
make db-migrate
make db-stop
make db-reset
make storage-start
make storage-stop
make mail-start
make mail-stop
make backend-run
make frontend-run
```

Локальная строка подключения:

```bash
postgres://forstitch:forstitch@localhost:5432/forstitch?sslmode=disable
```

## File Storage

Изображения товаров загружаются файлом через backend и хранятся в MinIO.

```bash
make storage-start
```

MinIO API доступен на `http://localhost:9000`, консоль — на `http://localhost:9001`.
Локальные учетные данные: `forstitch` / `forstitch-secret`.

## Dev Mail

Для локального перехвата писем используется Mailpit.

```bash
make mail-start
```

SMTP доступен на `localhost:1025`, веб-интерфейс для просмотра писем — `http://localhost:8025`.
Если backend будет запускаться внутри docker-сети, SMTP host нужно указывать как `mailpit`.
