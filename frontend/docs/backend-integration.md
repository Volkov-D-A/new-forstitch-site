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
    "img": "http://localhost:3000/api/files/products/lighthouse-aniva/1780000000000000000.jpg",
    "images": [
      { "id": 1, "url": "http://localhost:3000/api/files/products/lighthouse-aniva/1780000000000000001.jpg" }
    ],
    "isNew": true,
    "size": "300 x 220 крестов",
    "colors": "58 цветов DMC",
    "description": "Пейзажная схема с мягкими переходами и морским светом."
  }
]
```

- `id` используется в URL `/product/:productId` и в корзине.
- `cat` должен совпадать с `categories[].id`.
- `price` ожидается числом в рублях, форматирование делает фронтенд.
- `img` отдает backend после загрузки файла в админке; если поля нет, UI покажет placeholder.
- `isNew` вычисляет backend: новинками считаются последние 4 добавленных опубликованных товара.

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
    "excerpt": "Короткое описание публикации.",
    "content": "Полный текст публикации."
  }
]
```

Карточки блога показывают короткий `excerpt` и ведут на `/blog/:postId`, где отображается полный `content`.

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

Требует входа покупателя через `forstitch_customer_session`.

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
  "status": "paid",
  "message": "Заказ оформлен и считается оплаченным."
}
```

- Если покупатель не вошел, frontend перенаправит его в личный кабинет.
- После успешного заказа frontend покажет `message`, очистит корзину и закроет drawer.
- Если endpoint вернул ошибку, пользователь увидит toast, что оформление заказа пока не подключено.

Регистрация покупателя:

```http
POST /customer/register/start
POST /customer/register/verify
POST /customer/password-reset/start
POST /customer/password-reset/verify
POST /customer/login
GET  /customer/session
GET  /customer/orders
```

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
{ "username": "dimas", "password": "dimas" }
```

Backend выставляет HttpOnly cookie `forstitch_admin_session` и возвращает CSRF-токен:

```json
{ "username": "dimas", "csrfToken": "..." }
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
POST   /admin/products/:productId/image
DELETE /admin/products/:productId
GET    /admin/blog
POST   /admin/blog
PUT    /admin/blog/:postId
POST   /admin/blog/:postId/image
DELETE /admin/blog/:postId
GET    /admin/gallery
POST   /admin/gallery
PUT    /admin/gallery/:galleryItemId
POST   /admin/gallery/:galleryItemId/image
DELETE /admin/gallery/:galleryItemId
GET    /admin/site-settings
PUT    /admin/site-settings
GET    /admin/testimonials
POST   /admin/testimonials
PUT    /admin/testimonials/:testimonialId
POST   /admin/testimonials/:testimonialId/image
DELETE /admin/testimonials/:testimonialId
```

Payload для категорий:

```json
{ "label": "Цветы" }
```

При создании категорий и товаров `id` генерируется backend-ом и возвращается в ответе `POST`. Payload для товаров совпадает с элементом `GET /products`, но `id` при создании можно не передавать. Изображение товара не вводится URL вручную: админка отправляет `multipart/form-data` на `/admin/products/:productId/image` с полем `file`, а backend обновляет `img`.

Настройки главной страницы:

```json
{ "featuredProductId": "lighthouse_aniva" }
```

`featuredProductId` приходит в `GET /site-content` и определяет закрепленную схему на главном экране.

Отзывы вышивальщиц управляются из вкладки `Главная` в админке. Payload:

```json
{
  "name": "Анна",
  "role": "Вышивальщица",
  "img": "https://example.com/avatar.jpg",
  "text": "Очень понятная схема."
}
```

Фото отзыва не вводится URL вручную: админка отправляет `multipart/form-data` на `/admin/testimonials/:testimonialId/image` с полем `file`, а backend обновляет `img`.

Записи блога управляются из отдельной вкладки `Блог` в админке. При создании `id` генерируется backend-ом. Поле `excerpt` используется в карточке, `content` — на отдельной странице записи. Обложка не вводится URL вручную: админка отправляет `multipart/form-data` на `/admin/blog/:postId/image` с полем `file`, а backend обновляет `img`.

Галерея управляется из отдельной вкладки `Галерея` в админке. Изображение не вводится URL вручную: админка отправляет `multipart/form-data` на `/admin/gallery/:galleryItemId/image` с полем `file`, а backend обновляет `img`.

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
- Добавить серверную фильтрацию и пагинацию для большого каталога: `GET /products?category=&sort=&page=`.
- Расширить admin API на галерею, блог, site-content и заказы.
- При необходимости добавить runtime-схему валидации, например Zod, чтобы явно валидировать ответы API.
