CREATE TABLE IF NOT EXISTS tenants (
  id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  name TEXT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  ext_company_id BIGINT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE oauth_clients
  ADD CONSTRAINT fk_oauth_clients_company_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

ALTER TABLE oauth_auth_codes
  ADD CONSTRAINT fk_oauth_auth_codes_company_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

ALTER TABLE oauth_access_tokens
  ADD CONSTRAINT fk_oauth_access_tokens_company_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

ALTER TABLE oauth_refresh_tokens
  ADD CONSTRAINT fk_oauth_refresh_tokens_company_tenant
  FOREIGN KEY (company_id) REFERENCES tenants(id);

CREATE INDEX idx_oauth_clients_company ON oauth_clients(company_id);
CREATE INDEX idx_oauth_auth_codes_company ON oauth_auth_codes(company_id);
CREATE INDEX idx_oauth_access_tokens_company ON oauth_access_tokens(company_id);
CREATE INDEX idx_oauth_refresh_tokens_company ON oauth_refresh_tokens(company_id);
