CREATE TABLE customer_users (
  id bigserial PRIMARY KEY,
  email text NOT NULL UNIQUE,
  name text NOT NULL DEFAULT '',
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE customer_sessions (
  id text PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES customer_users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE orders
  ADD COLUMN customer_id bigint REFERENCES customer_users(id),
  ADD COLUMN customer_email text NOT NULL DEFAULT '',
  ADD COLUMN customer_name text NOT NULL DEFAULT '',
  ADD COLUMN email_token_hash text,
  ADD COLUMN email_token_expires_at timestamptz,
  ADD COLUMN email_confirmed_at timestamptz;

CREATE UNIQUE INDEX orders_email_token_hash_idx ON orders(email_token_hash) WHERE email_token_hash IS NOT NULL;
CREATE INDEX orders_customer_id_idx ON orders(customer_id);
