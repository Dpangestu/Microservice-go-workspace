CREATE TABLE IF NOT EXISTS user_profiles (
  user_id     CHAR(36)  NOT NULL,
  full_name   VARCHAR(255),
  display_name VARCHAR(255),
  phone       VARCHAR(32),
  avatar_url  TEXT,
  locale      VARCHAR(16),
  timezone    VARCHAR(64),
  metadata    JSON NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id),
  CONSTRAINT fk_user_profiles_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
