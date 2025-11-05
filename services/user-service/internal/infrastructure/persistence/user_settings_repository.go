package persistence

import (
	"database/sql"
	"log"

	"bkc_microservice/services/user-service/internal/domain/repositories"
)

type MySQLUserSettingsRepository struct {
	DB *sql.DB
}

func NewMySQLUserSettingsRepository(db *sql.DB) repositories.UserSettingsRepository {
	return &MySQLUserSettingsRepository{DB: db}
}

func (r *MySQLUserSettingsRepository) GetByUserID(userID string) (map[string]string, error) {
	rows, err := r.DB.Query(`
		SELECT k, v FROM user_settings
		WHERE user_id = ?
	`, userID)

	if err != nil {
		log.Printf("[UserSettingsRepo] Error fetching settings for user %s: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k string
		var v sql.NullString

		if err := rows.Scan(&k, &v); err != nil {
			log.Printf("[UserSettingsRepo] Error scanning row: %v", err)
			return nil, err
		}

		if v.Valid {
			settings[k] = v.String
		} else {
			settings[k] = ""
		}
	}

	return settings, nil
}

func (r *MySQLUserSettingsRepository) Set(userID, key, value string) error {
	_, err := r.DB.Exec(`
		INSERT INTO user_settings (id, user_id, k, v, created_at)
		VALUES (UUID(), ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE v = VALUES(v), updated_at = NOW()
	`, userID, key, value)
	return err
}

func (r *MySQLUserSettingsRepository) Delete(userID, key string) error {
	_, err := r.DB.Exec(`
		DELETE FROM user_settings
		WHERE user_id = ? AND k = ?
	`, userID, key)
	return err
}
