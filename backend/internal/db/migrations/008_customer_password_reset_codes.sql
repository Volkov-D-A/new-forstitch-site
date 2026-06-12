CREATE TABLE customer_password_reset_codes (
  email text PRIMARY KEY REFERENCES customer_users(email) ON DELETE CASCADE,
  code_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
