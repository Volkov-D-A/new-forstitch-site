CREATE TABLE customer_registration_codes (
  email text PRIMARY KEY,
  name text NOT NULL DEFAULT '',
  password_hash text NOT NULL,
  code_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
