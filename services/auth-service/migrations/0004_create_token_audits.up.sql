CREATE TABLE IF NOT EXISTS oauth_token_audits (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  company_id CHAR(36),
  client_id CHAR(36),
  user_id CHAR(36),
  token_type VARCHAR(50) NOT NULL,
  issued_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ip TEXT,
  user_agent TEXT,
  FOREIGN KEY (company_id) REFERENCES tenants(id),
  FOREIGN KEY (client_id) REFERENCES oauth_clients(id),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
