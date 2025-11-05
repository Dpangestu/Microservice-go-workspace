CREATE TABLE IF NOT EXISTS roles (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(255) NOT NULL,
  description TEXT,
  level       INT NOT NULL, -- Level menentukan hierarki role (admin, user, etc)
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO roles (name, description, level, is_active)
VALUES ('Default Role', 'Default role for existing users', 1, TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

ALTER TABLE roles
  ADD COLUMN tenant_id CHAR(36),
  ADD CONSTRAINT fk_roles_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE users
  ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(id)
  ON DELETE SET NULL ON UPDATE CASCADE;
