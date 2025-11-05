CREATE TABLE IF NOT EXISTS user_activities (
  id          CHAR(36) PRIMARY KEY DEFAULT (UUID()),
  user_id     CHAR(36) NOT NULL,
  action      VARCHAR(255) NOT NULL, -- Misalnya: "login", "update_profile"
  description TEXT,
  ip_address  VARCHAR(255),
  user_agent  TEXT,
  created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
