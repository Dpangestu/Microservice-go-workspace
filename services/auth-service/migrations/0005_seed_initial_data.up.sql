INSERT INTO tenants (id, name, status)
VALUES (UUID(), 'Default Tenant', 'active')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO oauth_clients (id, client_id, client_secret, redirect_uri, scopes, created_at)
VALUES (
  UUID(),
  'auth-service-client',
  'secret123',
  'http://localhost:8080/callback',
  'openid profile email',
  NOW()
)
ON DUPLICATE KEY UPDATE redirect_uri = VALUES(redirect_uri);

INSERT INTO users (
  id,
  username,
  email,
  password_hash,
  salt,
  role_id,
  is_active,
  is_locked,
  failed_login_attempts,
  last_login,
  last_password_change,
  password_expires_at,
  two_factor_enabled,
  two_factor_secret,
  session_token,
  remember_token,
  email_verified_at,
  phone_verified_at,
  created_at
)
VALUES (
  UUID(),
  'demolocal',
  'demo@bkc.local',
  '$2a$10$H6vB.1cOBeGHurYb1Um0eOOES4A5OcWxJp9o.wXaj6zErya1YAe3e',  -- hash dari "demo123"
  UUID(),
  1,
  TRUE,
  FALSE,
  0,
  NULL,
  NULL,
  NULL,
  FALSE,
  NULL,
  NULL,
  NULL,
  NULL,
  NULL,
  NOW()
)
ON DUPLICATE KEY UPDATE
  email = VALUES(email),
  password_hash = VALUES(password_hash);
