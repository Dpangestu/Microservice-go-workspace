-- ==============================================
-- 🚀 Initial Seeder for Auth & User Service
-- Compatible with existing schema (0001–0010)
-- ==============================================

START TRANSACTION;

-- ===== Tenants =====
INSERT INTO tenants (id, name, status, ext_company_id, created_at)
VALUES
  ('8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2', 'BKC Core Company', 'active', 1, NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ===== Roles =====
INSERT INTO roles (id, name, description, level, is_active, tenant_id, created_at)
VALUES
  (1, 'superadmin', 'Full system administrator', 100, TRUE, '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2', NOW()),
  (2, 'manager', 'Manager-level access', 80, TRUE, '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2', NOW()),
  (3, 'employee', 'Basic user role', 50, TRUE, '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2', NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ===== Permissions =====
INSERT INTO permissions (id, name, description, resource, action, created_at)
VALUES
  (1, 'user.view', 'View user data', 'users', 'read', NOW()),
  (2, 'user.manage', 'Manage user data', 'users', 'write', NOW()),
  (3, 'role.manage', 'Manage roles and permissions', 'roles', 'write', NOW()),
  (4, 'tenant.manage', 'Manage tenants', 'tenants', 'write', NOW()),
  (5, 'audit.view', 'View audit logs', 'audits', 'read', NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- ===== Role Permissions =====
INSERT INTO role_permissions (role_id, permission_id)
VALUES
  (1, 1),
  (1, 2),
  (1, 3),
  (1, 4),
  (1, 5),
  (2, 1),
  (2, 2),
  (3, 1)
ON DUPLICATE KEY UPDATE permission_id = VALUES(permission_id);

-- ===== Users =====
-- ===== Users =====
DELETE FROM users WHERE id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

INSERT INTO users (
  id, username, email, password_hash, salt, role_id, is_active, is_locked,
  failed_login_attempts, last_login, last_password_change, password_expires_at,
  two_factor_enabled, two_factor_secret, session_token, remember_token,
  email_verified_at, phone_verified_at, created_at
)
VALUES (
  'b262b15b-1b9b-4ed1-beb4-992c47a5943a',
  'admin',
  'admin@bkc.local',
  -- '$2a$12$OBl95mK9YfJh9WkA.bv81OqjFmR3GSp1xqIY/G6hBzv1IMQt/2ryS',
  '$2a$12$phxtCX.ZqFLYVAkd6R5OdevG7Zn6nBPcrpp/cRuD7OW8ogOxFWVPu',  -- password: admin123
  UUID(),
  1,
  TRUE,
  FALSE,
  0,
  NOW(),
  NOW(),
  NULL,
  FALSE,
  NULL,
  NULL,
  NULL,
  NOW(),
  NOW(),
  NOW()
);


-- ===== User Profile =====
-- Ensure user exists (for FK consistency)
SELECT id FROM users WHERE id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

INSERT INTO user_profiles (
  user_id, full_name, display_name, phone, avatar_url, locale, timezone, metadata, created_at
)
VALUES (
  'b262b15b-1b9b-4ed1-beb4-992c47a5943a',
  'System Administrator',
  'Admin BKC',
  '+628111111111',
  NULL,
  'id',
  'Asia/Jakarta',
  JSON_OBJECT('title', 'Super Admin', 'department', 'IT'),
  NOW()
)
ON DUPLICATE KEY UPDATE full_name = VALUES(full_name);

-- ===== User Settings =====
INSERT INTO user_settings (id, user_id, k, v)
VALUES
  (UUID(), 'b262b15b-1b9b-4ed1-beb4-992c47a5943a', 'theme', 'dark'),
  (UUID(), 'b262b15b-1b9b-4ed1-beb4-992c47a5943a', 'language', 'id')
ON DUPLICATE KEY UPDATE v = VALUES(v);

-- ===== OAuth Clients =====
INSERT INTO oauth_clients (id, client_id, client_secret, redirect_uri, scopes, company_id, created_at)
VALUES
  ('b26b8de5-4e71-4ecf-8321-8fa2e7c4443f',
   'bkc-core-web',
   '$2a$12$uZ0ZQr8D2a7cvjWq4bmvF.0mcOgr7ZbSY0/y3K6BfPjsyGXNUZrhy',  -- secret: web-secret
   'http://localhost:5173/callback',
   'openid profile email offline_access',
   '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2',
   NOW()),
  ('a5a5d3d3-91e4-4f23-b4cd-6f4dd74254e1',
   'bkc-mobile',
   '$2a$12$Y0QvFteNpgRHuAqf.MF7muuQkKczLRq0JknvGZai7zJChs9b74Tbe',  -- secret: mobile-secret
   'com.bkc.mobile://callback',
   'openid profile email offline_access',
   '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2',
   NOW())
ON DUPLICATE KEY UPDATE redirect_uri = VALUES(redirect_uri);

-- ===== Token Audit Example =====
INSERT INTO oauth_token_audits (id, company_id, client_id, user_id, token_type, issued_at, ip, user_agent)
VALUES (
  UUID(),
  '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2',
  'b26b8de5-4e71-4ecf-8321-8fa2e7c4443f',
  'b262b15b-1b9b-4ed1-beb4-992c47a5943a',
  'access_token',
  NOW(),
  '127.0.0.1',
  'Seeder Test Agent'
);

COMMIT;
