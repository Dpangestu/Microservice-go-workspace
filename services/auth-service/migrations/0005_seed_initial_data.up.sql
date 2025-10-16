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



INSERT INTO users (id, email, password_hash, created_at)
VALUES (
  UUID(),
  'demo@bkc.local',
  '$2a$10$H6vB.1cOBeGHurYb1Um0eOOES4A5OcWxJp9o.wXaj6zErya1YAe3e',  -- Password hash bcrypt untuk demo123
  NOW()
)
ON DUPLICATE KEY UPDATE email = VALUES(email), password_hash = VALUES(password_hash);