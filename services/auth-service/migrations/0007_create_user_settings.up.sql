CREATE TABLE IF NOT EXISTS user_settings (
  id         CHAR(36)  PRIMARY KEY DEFAULT (UUID()),
  user_id    CHAR(36)  NOT NULL,
  k          VARCHAR(100) NOT NULL,
  v          TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT uq_user_settings UNIQUE (user_id, k),
  CONSTRAINT fk_user_settings_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_settings_user ON user_settings(user_id);
