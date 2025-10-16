DROP INDEX IF EXISTS idx_oauth_tokens_company ON oauth_tokens;
DROP INDEX IF EXISTS idx_oauth_auth_codes_company ON oauth_auth_codes;
DROP INDEX IF EXISTS idx_oauth_clients_company ON oauth_clients;

ALTER TABLE oauth_tokens DROP FOREIGN KEY fk_oauth_tokens_tenant;
ALTER TABLE oauth_auth_codes DROP FOREIGN KEY fk_oauth_auth_codes_tenant;
ALTER TABLE oauth_clients DROP FOREIGN KEY fk_oauth_clients_tenant;

DROP TABLE IF EXISTS tenants;
