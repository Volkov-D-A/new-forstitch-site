# new-forstitch-site

Инструкция и примеры конфигурации для production-развертывания находятся в
[`Docs/README.md`](Docs/README.md).

Перед выпуском новой версии выполнить полный набор проверок:

```bash
make release-gate
```

Успешные проверки выводятся одной строкой с результатом `ok`. Если команда
завершается с ошибкой, выводится ее полный лог, а дальнейшие проверки не
запускаются.

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

## Локальные сервисы

```bash
docker compose up -d
```

Или через `Makefile`:

```bash
make services-start
make db-migrate
make services-stop
make services-reset
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
make services-start
```

MinIO API доступен на `http://localhost:9000`, консоль — на `http://localhost:9001`.
Локальные учетные данные: `forstitch` / `forstitch-secret`.

## Dev Mail

Для локального перехвата писем используется Mailpit.

```bash
make services-start
```

SMTP доступен на `localhost:1025`, веб-интерфейс для просмотра писем — `http://localhost:8025`.
Если backend будет запускаться внутри docker-сети, SMTP host нужно указывать как `mailpit`.
