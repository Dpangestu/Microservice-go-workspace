CREATE TABLE IF NOT EXISTS tenants (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  name TEXT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  ext_company_id BIGINT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tambah FK dari oauth_* ke tenants
ALTER TABLE oauth_clients
  ADD CONSTRAINT fk_oauth_clients_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

ALTER TABLE oauth_auth_codes
  ADD CONSTRAINT fk_oauth_auth_codes_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

ALTER TABLE oauth_tokens
  ADD CONSTRAINT fk_oauth_tokens_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

-- Index filter per tenant
CREATE INDEX idx_oauth_clients_company ON oauth_clients(company_id);
CREATE INDEX idx_oauth_auth_codes_company ON oauth_auth_codes(company_id);
CREATE INDEX idx_oauth_tokens_company ON oauth_tokens(company_id);
