-- oauth_clients
CREATE TABLE IF NOT EXISTS oauth_clients (
  id            CHAR(36)  PRIMARY KEY DEFAULT (UUID()),
  client_id     VARCHAR(255) UNIQUE NOT NULL,
  client_secret VARCHAR(255),
  redirect_uri  TEXT,
  scopes        TEXT,
  company_id    CHAR(36),
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- oauth_auth_codes
CREATE TABLE IF NOT EXISTS oauth_auth_codes (
  id                    CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  code                  VARCHAR(255) NOT NULL UNIQUE,
  user_id               CHAR(36),
  client_id             CHAR(36) NOT NULL,
  code_challenge        TEXT,
  code_challenge_method TEXT,
  redirect_uri          TEXT,
  scopes                TEXT,
  company_id            CHAR(36),
  expires_at            TIMESTAMP NOT NULL,
  created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id)   REFERENCES users(id),
  FOREIGN KEY (client_id) REFERENCES oauth_clients(id)
);

-- oauth_access_tokens
CREATE TABLE IF NOT EXISTS oauth_access_tokens (
  id           CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  user_id      CHAR(36),
  client_id    CHAR(36) NOT NULL,

  token        TEXT NOT NULL,
  token_sha    BINARY(32) AS (UNHEX(SHA2(token,256))) STORED,

  scopes       TEXT,
  company_id   CHAR(36),
  expires_at   TIMESTAMP NOT NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked      TINYINT(1) NOT NULL DEFAULT 0,

  FOREIGN KEY (user_id)   REFERENCES users(id),
  FOREIGN KEY (client_id) REFERENCES oauth_clients(id),

  UNIQUE KEY uq_oat_token_sha (token_sha),
  INDEX idx_oat_client (client_id),
  INDEX idx_oat_user (user_id),
  INDEX idx_oat_expires (expires_at)
);

-- oauth_refresh_tokens
CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
  id              CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  access_token_id CHAR(36) NOT NULL,

  token           TEXT NOT NULL,
  token_sha       BINARY(32) AS (UNHEX(SHA2(token,256))) STORED,

  company_id      CHAR(36),
  expires_at      TIMESTAMP NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked         TINYINT(1) NOT NULL DEFAULT 0,

  FOREIGN KEY (access_token_id) REFERENCES oauth_access_tokens(id) ON DELETE CASCADE,

  UNIQUE KEY uq_ort_token_sha (token_sha),
  INDEX idx_ort_expires (expires_at)
);
