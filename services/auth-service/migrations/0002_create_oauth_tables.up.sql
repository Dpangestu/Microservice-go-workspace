CREATE TABLE IF NOT EXISTS oauth_clients (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  client_id VARCHAR(255) UNIQUE NOT NULL,
  client_secret VARCHAR(255),
  redirect_uri TEXT,
  scopes TEXT,
  company_id CHAR(36),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS oauth_auth_codes (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  code VARCHAR(255) NOT NULL UNIQUE,
  user_id CHAR(36),
  client_id CHAR(36),
  code_challenge TEXT,
  code_challenge_method TEXT,
  redirect_uri TEXT,
  scopes TEXT,
  company_id CHAR(36),
  expires_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (client_id) REFERENCES oauth_clients(id)
);

CREATE TABLE IF NOT EXISTS oauth_tokens (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  user_id CHAR(36),
  client_id CHAR(36),
  access_token TEXT NOT NULL,
  refresh_token TEXT,
  scopes TEXT,
  company_id CHAR(36),
  expires_at TIMESTAMP NOT NULL,
  refresh_expires_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (client_id) REFERENCES oauth_clients(id)
);

-- Membuat indeks pada refresh_token dengan panjang terbatas
CREATE UNIQUE INDEX idx_oauth_tokens_refresh ON oauth_tokens(refresh_token);

-- Membuat indeks untuk kolom lainnya
CREATE INDEX idx_oauth_tokens_user ON oauth_tokens(user_id);
CREATE INDEX idx_oauth_tokens_client ON oauth_tokens(client_id);
CREATE INDEX idx_oauth_tokens_refresh_exp ON oauth_tokens(refresh_expires_at);

