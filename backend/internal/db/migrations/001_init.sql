CREATE TABLE categories (
  id text PRIMARY KEY,
  label text NOT NULL,
  sort_order integer NOT NULL DEFAULT 0
);

CREATE TABLE products (
  id text PRIMARY KEY,
  title text NOT NULL,
  price integer NOT NULL CHECK (price >= 0),
  cat text NOT NULL REFERENCES categories(id),
  sub text NOT NULL DEFAULT '',
  img text NOT NULL DEFAULT '',
  size text NOT NULL DEFAULT '',
  colors text NOT NULL DEFAULT '',
  canvas text NOT NULL DEFAULT '',
  sort_order integer NOT NULL DEFAULT 0,
  published boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE gallery_items (
  id bigserial PRIMARY KEY,
  img text NOT NULL,
  title text NOT NULL,
  by_name text NOT NULL,
  sort_order integer NOT NULL DEFAULT 0,
  published boolean NOT NULL DEFAULT true
);

CREATE TABLE blog_posts (
  id text PRIMARY KEY,
  title text NOT NULL,
  post_date date NOT NULL,
  tag text NOT NULL DEFAULT '',
  img text NOT NULL DEFAULT '',
  excerpt text NOT NULL DEFAULT '',
  published boolean NOT NULL DEFAULT true
);

CREATE TABLE site_content (
  id boolean PRIMARY KEY DEFAULT true CHECK (id),
  author_name text NOT NULL,
  author_photo text NOT NULL DEFAULT '',
  author_p1 text NOT NULL DEFAULT '',
  author_p2 text NOT NULL DEFAULT '',
  author_p3 text NOT NULL DEFAULT '',
  author_sign text NOT NULL DEFAULT ''
);

CREATE TABLE how_to_steps (
  n text PRIMARY KEY,
  title text NOT NULL,
  description text NOT NULL,
  sort_order integer NOT NULL DEFAULT 0
);

CREATE TABLE testimonials (
  id bigserial PRIMARY KEY,
  name text NOT NULL,
  role text NOT NULL DEFAULT '',
  img text NOT NULL DEFAULT '',
  text text NOT NULL,
  sort_order integer NOT NULL DEFAULT 0,
  published boolean NOT NULL DEFAULT true
);

CREATE TABLE orders (
  id text PRIMARY KEY,
  status text NOT NULL DEFAULT 'created',
  message text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
  order_id text NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id text NOT NULL REFERENCES products(id),
  quantity integer NOT NULL CHECK (quantity > 0),
  price integer NOT NULL CHECK (price >= 0),
  PRIMARY KEY (order_id, product_id)
);
