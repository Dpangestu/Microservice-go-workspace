package persistence

import (
	"database/sql"
	"encoding/json"
	"log"

	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
)

type MySQLUserProfileRepository struct {
	DB *sql.DB
}

func NewMySQLUserProfileRepository(db *sql.DB) repositories.UserProfileRepository {
	return &MySQLUserProfileRepository{DB: db}
}

func (r *MySQLUserProfileRepository) FindByUserID(userID string) (*entities.UserProfile, error) {
	row := r.DB.QueryRow(`
		SELECT 
			user_id, full_name, display_name, phone, avatar_url, locale, timezone, metadata, created_at, updated_at
		FROM user_profiles
		WHERE user_id = ?
	`, userID)

	profile := &entities.UserProfile{}
	var metadata sql.NullString

	err := row.Scan(
		&profile.UserID, &profile.FullName, &profile.DisplayName, &profile.Phone,
		&profile.AvatarURL, &profile.Locale, &profile.Timezone, &metadata,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Profile tidak ada, kembalikan nil (bukan error)
	}
	if err != nil {
		log.Printf("[UserProfileRepo] Error fetching profile for user %s: %v", userID, err)
		return nil, err
	}

	if metadata.Valid {
		profile.Metadata = json.RawMessage(metadata.String)
	}

	return profile, nil
}

func (r *MySQLUserProfileRepository) Create(profile *entities.UserProfile) error {
	_, err := r.DB.Exec(`
		INSERT INTO user_profiles (user_id, full_name, display_name, phone, avatar_url, locale, timezone, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`, profile.UserID, profile.FullName, profile.DisplayName, profile.Phone,
		profile.AvatarURL, profile.Locale, profile.Timezone, profile.Metadata)
	return err
}

func (r *MySQLUserProfileRepository) Update(profile *entities.UserProfile) error {
	_, err := r.DB.Exec(`
		UPDATE user_profiles 
		SET full_name = ?, display_name = ?, phone = ?, avatar_url = ?, locale = ?, timezone = ?, metadata = ?, updated_at = NOW()
		WHERE user_id = ?
	`, profile.FullName, profile.DisplayName, profile.Phone, profile.AvatarURL,
		profile.Locale, profile.Timezone, profile.Metadata, profile.UserID)
	return err
}
