CREATE TABLE admin_users (
  id bigserial PRIMARY KEY,
  username text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE admin_sessions (
  id text PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
  csrf_token text NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX admin_sessions_user_id_idx ON admin_sessions(user_id);
CREATE INDEX admin_sessions_expires_at_idx ON admin_sessions(expires_at);
