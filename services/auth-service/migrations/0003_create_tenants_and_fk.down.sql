ALTER TABLE oauth_refresh_tokens  DROP FOREIGN KEY fk_oauth_refresh_tokens_tenant;
ALTER TABLE oauth_access_tokens   DROP FOREIGN KEY fk_oauth_access_tokens_tenant;
ALTER TABLE oauth_auth_codes      DROP FOREIGN KEY fk_oauth_auth_codes_tenant;
ALTER TABLE oauth_clients         DROP FOREIGN KEY fk_oauth_clients_tenant;

DROP INDEX idx_oauth_refresh_tokens_company  ON oauth_refresh_tokens;
DROP INDEX idx_oauth_access_tokens_company   ON oauth_access_tokens;
DROP INDEX idx_oauth_auth_codes_company      ON oauth_auth_codes;
DROP INDEX idx_oauth_clients_company         ON oauth_clients;

DROP TABLE IF EXISTS tenants;
