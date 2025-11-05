CREATE TABLE IF NOT EXISTS sycrone_core (
  id              CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  user_id         CHAR(36) NOT NULL,
  user_core       CHAR(36) NOT NULL,
  kode_group_1    VARCHAR(100) NOT NULL,
  kode_perkiraan  VARCHAR(100) NOT NULL DEFAULT '10102',
  created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
