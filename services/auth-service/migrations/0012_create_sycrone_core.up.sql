CREATE TABLE IF NOT EXISTS sycrone_core (
  id              CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  user_id         CHAR(36) NOT NULL UNIQUE,  -- One-to-one with users
  user_core       VARCHAR(100) NOT NULL UNIQUE,  -- CBS user ID (might not be UUID)
  kode_group_1    VARCHAR(100) NOT NULL,
  kode_perkiraan  VARCHAR(100) NOT NULL DEFAULT '10102',
  kode_cabang     VARCHAR(50),   -- Branch code (optional)
  status          ENUM('active', 'suspended', 'inactive') DEFAULT 'active',
  sync_status     ENUM('synced', 'pending', 'failed') DEFAULT 'pending',
  last_sync_at    TIMESTAMP,
  sync_error      TEXT,
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_core (user_core),
  INDEX idx_status (status),
  INDEX idx_sync_status (sync_status)
);
