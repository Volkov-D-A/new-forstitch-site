INSERT INTO categories (id, label, sort_order) VALUES
  ('fauna', 'Животный мир', 10),
  ('people', 'Люди', 20),
  ('still-life', 'Натюрморты', 30),
  ('landscape', 'Пейзаж', 40),
  ('fantasy', 'Фэнтези', 50)
ON CONFLICT (id) DO NOTHING;

INSERT INTO products (
  id, title, price, cat, sub, img, is_new, size, colors, canvas, sort_order
) VALUES
  (
    'lighthouse_aniva',
    'Маяк на мысе Анива',
    600,
    'landscape',
    'Море',
    'https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg',
    true,
    '300 x 220 крестов',
    '58 цветов DMC',
    'Aida 16 / равномерка 32',
    10
  ),
  (
    'oxota_na_miod',
    'Охота на мед',
    200,
    'fauna',
    'Насекомые',
    'https://forstitch.ru/wp-content/uploads/2021/04/5-300x300.jpg',
    true,
    '120 x 120 крестов',
    '32 цвета DMC',
    'Aida 14',
    20
  ),
  (
    'dragon_library',
    'Дракон-читальня',
    450,
    'fantasy',
    'Драконы',
    'https://forstitch.ru/wp-content/uploads/2016/11/8SNwJDfXaw-1030x833.jpg',
    false,
    '240 x 190 крестов',
    '52 цвета DMC',
    'Aida 16',
    30
  ),
  (
    'anemones',
    'Анемоны',
    400,
    'still-life',
    'Цветы',
    'https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg',
    false,
    '180 x 240 крестов',
    '46 цветов',
    'равномерка 32',
    40
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO gallery_items (img, title, by_name, sort_order) VALUES
  ('https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg', 'Маяк на мысе Анива', 'Команда Forstitch', 10),
  ('https://forstitch.ru/wp-content/uploads/2016/11/oQrdgtvEwgs-773x1030.jpg', 'Анемоны', 'Команда Forstitch', 20)
ON CONFLICT DO NOTHING;

INSERT INTO blog_posts (id, title, post_date, tag, img, excerpt) VALUES
  (
    'new-patterns',
    'Новые схемы в каталоге',
    '2026-06-10',
    'Новости',
    'https://forstitch.ru/wp-content/uploads/2021/05/16-495x400.jpg',
    'Первые товары уже отдаются из PostgreSQL. Дальше сюда можно подключить админку.'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO site_content (
  id, author_name, author_photo, author_p1, author_p2, author_p3, author_sign
) VALUES (
  true,
  'Екатерина Волкова',
  'https://forstitch.ru/wp-content/uploads/2016/04/MG_4272-687x1030.jpg',
  'Авторские схемы для вышивки крестом с вниманием к цвету, деталям и удобству отшива.',
  'Каждая схема готовится вручную и проверяется перед публикацией.',
  'Сайт постепенно переезжает на новый backend, чтобы каталогом было удобно управлять.',
  'Екатерина'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO how_to_steps (n, title, description, sort_order) VALUES
  ('01', 'Выберите схему', 'Добавьте понравившуюся PDF-схему в корзину.', 10),
  ('02', 'Оформите заказ', 'Backend создаст заказ и подготовит ссылку на оплату.', 20),
  ('03', 'Оплатите', 'После оплаты схема будет отправлена на указанную почту.', 30),
  ('04', 'Вышивайте', 'Откройте PDF-файл, подготовьте материалы и начинайте отшив.', 40)
ON CONFLICT (n) DO NOTHING;

INSERT INTO testimonials (name, role, img, text, sort_order) VALUES
  (
    'Мария',
    'Вышивальщица',
    'https://forstitch.ru/wp-content/uploads/2021/04/5-300x300.jpg',
    'Плавные переходы и понятная схема, приятно вышивать.',
    10
  )
ON CONFLICT DO NOTHING;
