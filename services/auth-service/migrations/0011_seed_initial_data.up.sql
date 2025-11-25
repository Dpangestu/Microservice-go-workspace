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
  (1, 'user.read', 'View user list and details', 'users', 'read', NOW()),
  (2, 'user.read_self', 'View own profile only', 'users', 'read_self', NOW()),
  (3, 'user.create', 'Create new users', 'users', 'create', NOW()),
  (4, 'user.update', 'Update user information', 'users', 'update', NOW()),
  (5, 'user.update_self', 'Update own profile only', 'users', 'update_self', NOW()),
  (6, 'user.delete', 'Delete users', 'users', 'delete', NOW()),
  (7, 'user.manage_profile', 'Manage user profiles', 'users', 'manage_profile', NOW()),
  (8, 'user.manage_settings', 'Manage user settings', 'users', 'manage_settings', NOW()),
  (9, 'user.view_activity', 'View user activity logs', 'users', 'view_activity', NOW()),

-- 3.2 Roles Permissions (4)
  (10, 'role.read', 'View roles', 'roles', 'read', NOW()),
  (11, 'role.create', 'Create new roles', 'roles', 'create', NOW()),
  (12, 'role.update', 'Update roles', 'roles', 'update', NOW()),
  (13, 'role.delete', 'Delete roles', 'roles', 'delete', NOW()),

-- 3.3 Permissions Management (4)
  (14, 'permission.read', 'View permissions', 'permissions', 'read', NOW()),
  (15, 'permission.create', 'Create permissions', 'permissions', 'create', NOW()),
  (16, 'permission.update', 'Update permissions', 'permissions', 'update', NOW()),
  (17, 'permission.delete', 'Delete permissions', 'permissions', 'delete', NOW()),

-- 3.4 Role Permissions (2)
  (18, 'role_permission.assign', 'Assign permissions to roles', 'role_permissions', 'assign', NOW()),
  (19, 'role_permission.revoke', 'Revoke permissions from roles', 'role_permissions', 'revoke', NOW()),

-- 3.5 Audit & Logs (4)
  (20, 'audit.read', 'View audit logs', 'audits', 'read', NOW()),
  (21, 'audit.export', 'Export audit logs', 'audits', 'export', NOW()),
  (22, 'log.read', 'View system logs', 'logs', 'read', NOW()),
  (23, 'log.delete', 'Delete logs', 'logs', 'delete', NOW()),

-- 3.6 Reports (3)
  (24, 'report.read', 'View reports', 'reports', 'read', NOW()),
  (25, 'report.create', 'Create reports', 'reports', 'create', NOW()),
  (26, 'report.export', 'Export reports', 'reports', 'export', NOW()),

-- 3.7 Tenants (5)
  (27, 'tenant.read', 'View tenants', 'tenants', 'read', NOW()),
  (28, 'tenant.create', 'Create new tenants', 'tenants', 'create', NOW()),
  (29, 'tenant.update', 'Update tenant information', 'tenants', 'update', NOW()),
  (30, 'tenant.delete', 'Delete tenants', 'tenants', 'delete', NOW()),
  (31, 'tenant.manage_settings', 'Manage tenant settings', 'tenants', 'manage_settings', NOW()),

-- 3.8 System Configuration (4)
  (32, 'system.read_config', 'Read system configuration', 'system', 'read_config', NOW()),
  (33, 'system.write_config', 'Modify system configuration', 'system', 'write_config', NOW()),
  (34, 'system.manage_backup', 'Manage backups', 'system', 'manage_backup', NOW()),
  (35, 'system.manage_database', 'Manage database', 'system', 'manage_database', NOW()),

-- 3.9 Features (2)
  (36, 'feature.read', 'View available features', 'features', 'read', NOW()),
  (37, 'feature.enable', 'Enable/disable features', 'features', 'enable', NOW()),

-- 3.10 Settings (2)
  (38, 'setting.read', 'Read settings', 'settings', 'read', NOW()),
  (39, 'setting.write', 'Write settings', 'settings', 'write', NOW()),

-- 3.11 API Keys (3)
  (40, 'api_key.read', 'View API keys', 'api_keys', 'read', NOW()),
  (41, 'api_key.create', 'Create API keys', 'api_keys', 'create', NOW()),
  (42, 'api_key.delete', 'Revoke API keys', 'api_keys', 'delete', NOW()),

-- 3.12 Webhooks (4)
  (43, 'webhook.read', 'View webhooks', 'webhooks', 'read', NOW()),
  (44, 'webhook.create', 'Create webhooks', 'webhooks', 'create', NOW()),
  (45, 'webhook.update', 'Update webhooks', 'webhooks', 'update', NOW()),
  (46, 'webhook.delete', 'Delete webhooks', 'webhooks', 'delete', NOW()),

-- 3.13 Dashboard & Analytics (3)
  (47, 'dashboard.read', 'View dashboard', 'dashboard', 'read', NOW()),
  (48, 'analytics.read', 'View analytics', 'analytics', 'read', NOW()),
  (49, 'notification.read', 'Read notifications', 'notifications', 'read', NOW()),

-- 3.14 Notifications (2)
  (50, 'notification.manage', 'Manage notifications', 'notifications', 'manage', NOW()),

-- 3.15 Content Management (5)
  (51, 'content.read', 'View content', 'content', 'read', NOW()),
  (52, 'content.create', 'Create content', 'content', 'create', NOW()),
  (53, 'content.update', 'Update content', 'content', 'update', NOW()),
  (54, 'content.delete', 'Delete content', 'content', 'delete', NOW()),
  (55, 'content.publish', 'Publish content', 'content', 'publish', NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- ===== Role Permissions =====
INSERT INTO role_permissions (role_id, permission_id)
VALUES
  (1,1),(1,2),(1,3),(1,4),(1,5),(1,6),(1,7),(1,8),(1,9),
  (1,10),(1,11),(1,12),(1,13),(1,14),(1,15),(1,16),(1,17),(1,18),(1,19),
  (1,20),(1,21),(1,22),(1,23),(1,24),(1,25),(1,26),(1,27),(1,28),(1,29),(1,30),(1,31),
  (1,32),(1,33),(1,34),(1,35),(1,36),(1,37),(1,38),(1,39),(1,40),(1,41),(1,42),
  (1,43),(1,44),(1,45),(1,46),(1,47),(1,48),(1,49),(1,50),(1,51),(1,52),(1,53),(1,54),(1,55)
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
