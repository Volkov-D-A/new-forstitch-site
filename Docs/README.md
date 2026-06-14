# Production deployment

Эта папка содержит пример развертывания Forstitch на одном Ubuntu-сервере.
На сервере нужны только Docker Engine, Docker Compose plugin и файлы из этой
папки.

## Архитектура

```text
Internet
  |
  +-- :80/:443 -> Caddy
                     |
                     +-- /api/* -> backend:3000
                     +-- /*     -> frontend:80

backend -> postgres:5432
backend -> minio:9000
backend -> SMTP provider
```

Наружу публикуются только порты `80` и `443`. PostgreSQL, MinIO, backend и
frontend доступны только внутри Docker-сети.

Caddy автоматически получает и продлевает TLS-сертификат для домена через
Let's Encrypt или другой поддерживаемый ACME-сервис. Данные сертификатов
хранятся в Docker volume `caddy_data`.

## Состав

- `docker-compose.prod.yml` — production-сервисы.
- `.env.prod.example` — шаблон production-переменных.
- `Caddyfile` — HTTPS, reverse proxy и маршрутизация.

Настоящий `.env.prod` должен храниться только на сервере и не должен попадать
в Git или Docker-образы.

## Требования к образам

До первого развертывания необходимо опубликовать в Docker Hub два образа:

```text
DOCKERHUB_USER/forstitch-backend:APP_VERSION
DOCKERHUB_USER/forstitch-frontend:APP_VERSION
```

Backend-образ должен содержать скомпилированное Go-приложение и запускать API
на `:3000`. SQL-миграции уже встраиваются в бинарник и применяются при старте.

В репозитории backend-образ собирается файлом `backend/Dockerfile`. Для
публикации указать в корневом `.env`:

```dotenv
DOCKERHUB_USER=your-dockerhub-user
APP_VERSION=1.0.0
```

Войти в Docker Hub и отправить multi-platform образ:

```bash
docker login
make backend-image-push
```

Будет опубликован образ:

```text
DOCKERHUB_USER/forstitch-backend:APP_VERSION
```

Локальная сборка без отправки выполняется командой:

```bash
make backend-image-build
```

Frontend-образ собирается файлом `frontend/Dockerfile`. Он содержит
production-сборку React в nginx и поддерживает SPA fallback на `index.html`.
Frontend обращается к относительному адресу `/api`, поэтому один образ работает
на любом домене без пересборки.

Собрать frontend-образ локально:

```bash
make frontend-image-build
```

Собрать multi-platform образ и отправить его в Docker Hub:

```bash
docker login
make frontend-image-push
```

Будет опубликован образ:

```text
DOCKERHUB_USER/forstitch-frontend:APP_VERSION
```

## Подготовка домена

1. Создать DNS-запись `A` для `forstitch.ru`, указывающую на IPv4 сервера.
2. При использовании IPv6 создать корректную запись `AAAA` или удалить ее.
3. Открыть входящие TCP-порты `80` и `443`, а также UDP-порт `443`.
4. Убедиться, что эти порты не заняты другим nginx, Apache или reverse proxy.

Порт `80` нужен в том числе для первичной проверки домена центром сертификации.

## Первый запуск

Скопировать файлы из `Docs` в отдельный каталог на сервере, например:

```bash
mkdir -p /opt/forstitch
cd /opt/forstitch
```

Создать production-конфигурацию:

```bash
cp .env.prod.example .env.prod
nano .env.prod
```

Обязательно заменить:

- `DOCKERHUB_USER`;
- `APP_VERSION`;
- `APP_DOMAIN`;
- теги инфраструктурных образов, если в шаблоне используется плавающий тег;
- пароли PostgreSQL, MinIO и администратора;
- настройки SMTP.

Значение `POSTGRES_PASSWORD` продублировано внутри `DATABASE_URL`. Если пароль
содержит `@`, `:`, `/`, `?`, `#` или другие специальные символы URL, в
`DATABASE_URL` их нужно записать в percent-encoded виде.

Если образы Docker Hub приватные, предварительно выполнить:

```bash
docker login
```

Проверить итоговую конфигурацию и запустить сервисы:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml config
docker compose --env-file .env.prod -f docker-compose.prod.yml pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
```

Проверка состояния:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
docker compose --env-file .env.prod -f docker-compose.prod.yml logs -f caddy backend
```

После успешного запуска сайт должен открываться по адресу из `APP_BASE_URL`.

## Обновление

Собрать и отправить в Docker Hub новую версию backend и frontend, затем изменить
`APP_VERSION` в `.env.prod`:

```dotenv
APP_VERSION=1.1.0
```

Применить обновление:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d
docker image prune
```

Не использовать `latest` для собственных production-образов. Версия должна
однозначно определять содержимое образа.

## Остановка

Остановить контейнеры без удаления данных:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml stop
```

Удалить контейнеры и сеть, сохранив volumes:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml down
```

Команда `down -v` удалит базу данных, файлы MinIO и сертификаты Caddy. В
production ее нельзя выполнять без осознанного решения и актуальной резервной
копии.

## Резервные копии

Минимально необходимо регулярно сохранять:

- дамп PostgreSQL;
- содержимое MinIO;
- `.env.prod`;
- volume `caddy_data` или позволить Caddy выпустить сертификаты заново.

Пример дампа PostgreSQL:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml \
  exec -T postgres pg_dump -U forstitch -d forstitch > postgres.sql
```

Имена пользователя и базы в команде должны соответствовать `.env.prod`.
Резервные копии нужно хранить вне этого сервера и периодически проверять
процедуру восстановления.

## Безопасность

- Не публиковать порты PostgreSQL, MinIO, backend и MinIO Console.
- Использовать уникальные длинные пароли.
- Ограничить SSH по ключам и настроить firewall.
- Регулярно обновлять Ubuntu, Docker и базовые образы.
- Не хранить секреты в Git, CI-логах и Dockerfile.
- Для production использовать реальный SMTP; Mailpit предназначен только для
  локальной разработки.

## Диагностика HTTPS

Проверить DNS:

```bash
getent hosts forstitch.ru
```

Посмотреть журнал Caddy:

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml logs caddy
```

Частые причины, по которым сертификат не выпускается:

- DNS еще указывает не на тот сервер;
- порты `80` или `443` закрыты firewall или панелью хостинга;
- порт занят другим процессом;
- существует некорректная запись `AAAA`;
- volume Caddy недоступен для записи.
