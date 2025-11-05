-- ==============================================
-- 🧹 Rollback Seeder for Auth & User Service
-- ==============================================

START TRANSACTION;

-- Delete Token Audits
DELETE FROM oauth_token_audits
WHERE user_id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

-- Delete OAuth Clients
DELETE FROM oauth_clients
WHERE id IN (
  'b26b8de5-4e71-4ecf-8321-8fa2e7c4443f',
  'a5a5d3d3-91e4-4f23-b4cd-6f4dd74254e1'
);

-- Delete User Settings
DELETE FROM user_settings
WHERE user_id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

-- Delete User Profiles
DELETE FROM user_profiles
WHERE user_id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

-- Delete Users
DELETE FROM users
WHERE id = 'b262b15b-1b9b-4ed1-beb4-992c47a5943a';

-- Delete Role Permissions
DELETE FROM role_permissions
WHERE role_id IN (1, 2, 3);

-- Delete Permissions
DELETE FROM permissions
WHERE id BETWEEN 1 AND 5;

-- Delete Roles
DELETE FROM roles
WHERE id IN (1, 2, 3);

-- Delete Tenants
DELETE FROM tenants
WHERE id = '8a6b0ad8-9f10-4f61-8f3d-9b7c1cf2e9a2';

COMMIT;
