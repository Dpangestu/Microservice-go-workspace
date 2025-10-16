-- +migrate Down
DELETE FROM oauth_clients WHERE client_id = 'auth-service-client';
DELETE FROM users WHERE email = 'admin@local.test';
DELETE FROM tenants WHERE name = 'Default Tenant';
