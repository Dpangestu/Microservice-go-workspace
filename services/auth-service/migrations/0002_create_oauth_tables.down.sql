DROP INDEX IF EXISTS idx_oauth_tokens_refresh_exp ON oauth_tokens;
DROP INDEX IF EXISTS idx_oauth_tokens_client ON oauth_tokens;
DROP INDEX IF EXISTS idx_oauth_tokens_user ON oauth_tokens;
DROP INDEX IF EXISTS idx_oauth_tokens_refresh ON oauth_tokens;

DROP TABLE IF EXISTS oauth_tokens;
DROP TABLE IF EXISTS oauth_auth_codes;
DROP TABLE IF EXISTS oauth_clients;
