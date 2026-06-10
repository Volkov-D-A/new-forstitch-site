# Подготовка фронтенда к подключению бэкенда

## Текущее состояние

- `src/services/siteApi.ts` — единая точка общения с API.
- `src/hooks/useSiteData.ts` — загрузка данных сайта, состояния loading/error.
- `src/types/site.ts` — общий контракт UI и API.
- `src/hooks/useCart.ts` — клиентская корзина в формате позиций `{ productId, quantity }`.

Фронтенд всегда обращается к Go API по адресу `http://localhost:3000/api`.

## API URL

Текущий базовый URL захардкожен в `src/services/siteApi.ts`. Фронтенд запрашивает:

```http
GET http://localhost:3000/api/categories
GET http://localhost:3000/api/products
GET http://localhost:3000/api/gallery
GET http://localhost:3000/api/blog
GET http://localhost:3000/api/site-content
POST http://localhost:3000/api/orders
```

## Endpoints

### `GET /categories`

```json
[
  { "id": "fauna", "label": "Животный мир" },
  { "id": "landscape", "label": "Пейзаж" }
]
```

- `id` — стабильный slug категории.
- `label` — отображаемое имя.
- Категорию `all` можно не отдавать: фронтенд добавит ее сам.

### `GET /products`

```json
[
  {
    "id": "lighthouse_aniva",
    "title": "Маяк на мысе Анива",
    "price": 600,
    "cat": "landscape",
    "sub": "Море",
    "img": "https://example.com/images/aniva.jpg",
    "isNew": true,
    "size": "300 x 220 крестов",
    "colors": "58 цветов DMC",
    "canvas": "Aida 16 / равномерка 32"
  }
]
```

- `id` используется в URL `/product/:productId` и в корзине.
- `cat` должен совпадать с `categories[].id`.
- `price` ожидается числом в рублях, форматирование делает фронтенд.
- `img` лучше отдавать абсолютным URL; если поля нет, UI покажет placeholder.

### `GET /products/:productId`

Возвращает один товар в той же форме, что элемент `GET /products`.

Сейчас страница товара ищет товар в уже загруженном списке, но endpoint оставлен в API-слое для следующего шага: отдельной загрузки товара по глубокой ссылке.

### `GET /gallery`

```json
[
  {
    "img": "https://example.com/gallery/work.jpg",
    "title": "Маяк на мысе Анива",
    "by": "Анна"
  }
]
```

### `GET /blog`

```json
[
  {
    "id": "new-patterns",
    "title": "Новые схемы",
    "date": "2026-06-10",
    "tag": "Новости",
    "img": "https://example.com/blog/post.jpg",
    "excerpt": "Короткое описание публикации."
  }
]
```

### `GET /site-content`

Endpoint обязателен для текущего фронтенда.

```json
{
  "author": {
    "name": "Екатерина Волкова",
    "photo": "https://example.com/author.jpg",
    "p1": "Первый абзац.",
    "p2": "Второй абзац.",
    "p3": "Третий абзац.",
    "sign": "Екатерина"
  },
  "howToBuy": [
    { "n": "01", "t": "Выберите схему", "d": "Добавьте PDF-схему в корзину." }
  ],
  "testimonials": [
    {
      "name": "Мария",
      "role": "Вышивальщица",
      "img": "https://example.com/reviews/maria.jpg",
      "text": "Текст отзыва."
    }
  ]
}
```

### `POST /orders`

Запрос:

```json
{
  "items": [
    { "productId": "lighthouse_aniva", "quantity": 1 }
  ]
}
```

Ответ:

```json
{
  "id": "order_123",
  "checkoutUrl": "https://pay.example.com/order_123",
  "message": "Заказ создан"
}
```

- Если `checkoutUrl` есть, фронтенд перенаправит пользователя на оплату.
- Если `checkoutUrl` нет, фронтенд покажет `message`, очистит корзину и закроет drawer.
- Если endpoint вернул ошибку, пользователь увидит toast, что оформление заказа пока не подключено.

## Что уже готово

- Категории больше не зашиты в TypeScript union и могут приходить из админки.
- Товары, категории и часть товарных полей проходят мягкую нормализацию.
- Корзина хранит позиции с количеством.
- Checkout имеет фронтовой контракт `createOrder()`.
- Фронтенд не содержит локальных тестовых данных и зависит от запущенного backend.

## Следующие шаги

## Admin API

Админка использует session-based auth:

```http
POST /auth/login
GET  /auth/session
POST /auth/logout
```

`POST /auth/login` принимает:

```json
{ "username": "admin", "password": "dev-admin-password" }
```

Backend выставляет HttpOnly cookie `forstitch_admin_session` и возвращает CSRF-токен:

```json
{ "username": "admin", "csrfToken": "..." }
```

Фронтенд отправляет admin-запросы с `credentials: "include"`. Для `POST`, `PUT`, `DELETE` нужен заголовок:

```http
X-CSRF-Token: <csrfToken>
```

```http
GET    /admin/categories
POST   /admin/categories
PUT    /admin/categories/:categoryId
DELETE /admin/categories/:categoryId
GET    /admin/products
POST   /admin/products
PUT    /admin/products/:productId
DELETE /admin/products/:productId
```

Payload для категорий:

```json
{ "id": "flowers", "label": "Цветы" }
```

Payload для товаров совпадает с элементом `GET /products`.

## Ошибки API

Backend возвращает ошибки в едином формате:

```json
{
  "error": {
    "code": "order_empty",
    "message": "order must contain at least one item"
  }
}
```

Фронтенду стоит ориентироваться на `error.code` для ветвления логики и показывать `error.message` как fallback-текст.

## Следующие шаги

- Подключить отдельную загрузку товара через `GET /products/:productId` на странице карточки.
- Добавить форму контактов в checkout, если заказ должен собирать email до оплаты.
- Добавить серверную фильтрацию и пагинацию для большого каталога: `GET /products?category=&sort=&page=`.
- Расширить admin API на галерею, блог, site-content и заказы.
- При необходимости добавить runtime-схему валидации, например Zod, чтобы явно валидировать ответы API.
