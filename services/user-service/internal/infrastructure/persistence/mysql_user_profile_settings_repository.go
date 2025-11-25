package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

type MySQLUserProfileRepository struct {
	db *sql.DB
}

func NewMySQLUserProfileRepository(db *sql.DB) *MySQLUserProfileRepository {
	return &MySQLUserProfileRepository{db: db}
}

// func (r *MySQLUserProfileRepository) FindByUserID(userID string) (*entities.UserProfile, error) {
// 	query := `
// 		SELECT id, user_id, first_name, last_name, phone, avatar, created_at, updated_at
// 		FROM user_profiles
// 		WHERE user_id = ?
// 	`

// 	profile := &entities.UserProfile{}
// 	err := r.db.QueryRow(query, userID).Scan(
// 		&profile.UserID, &profile.FullName, &profile.DisplayName, &profile.Phone,
// 		&profile.AvatarURL, &profile.Locale, &profile.Timezone, &metadata,
// 		&profile.CreatedAt, &profile.UpdatedAt,
// 	)

// 	if err == sql.ErrNoRows {
// 		return nil, fmt.Errorf("user profile not found")
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query profile: %w", err)
// 	}

// 	return profile, nil
// }

func (r *MySQLUserProfileRepository) FindByUserID(userID string) (*entities.UserProfile, error) {
	row := r.db.QueryRow(`
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

func (r *MySQLUserProfileRepository) Create(ctx context.Context, profile *entities.UserProfile) error {
	if profile == nil {
		return fmt.Errorf("profile is required")
	}

	query := `
		INSERT INTO user_profiles (user_id, first_name, last_name, phone, avatar, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		&profile.UserID, &profile.FullName, &profile.DisplayName, &profile.Phone,
		&profile.AvatarURL, &profile.Locale, &profile.Timezone,
		&profile.CreatedAt, &profile.UpdatedAt,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	profile.ID = int(id)
	return nil
}

func (r *MySQLUserProfileRepository) Update(ctx context.Context, profile *entities.UserProfile) error {
	if profile == nil {
		return fmt.Errorf("profile is required")
	}

	query := `
		UPDATE user_profiles
		SET first_name = ?, last_name = ?, phone = ?, avatar = ?, updated_at = ?
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		&profile.UserID, &profile.FullName, &profile.DisplayName, &profile.Phone,
		&profile.AvatarURL, &profile.Locale, &profile.Timezone,
		&profile.CreatedAt, &profile.UpdatedAt,
		time.Now(),
		profile.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

func (r *MySQLUserProfileRepository) Delete(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	query := "DELETE FROM user_profiles WHERE user_id = ?"

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user profile not found")
	}

	return nil
}

// MySQLUserSettingsRepository implements UserSettingsRepository interface
type MySQLUserSettingsRepository struct {
	db *sql.DB
}

func NewMySQLUserSettingsRepository(db *sql.DB) *MySQLUserSettingsRepository {
	return &MySQLUserSettingsRepository{db: db}
}

func (r *MySQLUserSettingsRepository) FindByUserID(userID string) (*entities.UserSettings, error) {
	query := `
		SELECT id, user_id, theme_mode, language, two_fa_enabled, notifications, created_at, updated_at
		FROM user_settings
		WHERE user_id = ?
	`

	settings := &entities.UserSettings{}
	err := r.db.QueryRow(query, userID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.ThemeMode,
		&settings.Language,
		&settings.TwoFAEnabled,
		&settings.Notifications,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user settings not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}

	return settings, nil
}

func (r *MySQLUserSettingsRepository) Create(ctx context.Context, settings *entities.UserSettings) error {
	if settings == nil {
		return fmt.Errorf("settings is required")
	}

	query := `
		INSERT INTO user_settings (user_id, theme_mode, language, two_fa_enabled, notifications, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		settings.UserID,
		settings.ThemeMode,
		settings.Language,
		settings.TwoFAEnabled,
		settings.Notifications,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	settings.ID = int(id)
	return nil
}

func (r *MySQLUserSettingsRepository) Update(ctx context.Context, settings *entities.UserSettings) error {
	if settings == nil {
		return fmt.Errorf("settings is required")
	}

	query := `
		UPDATE user_settings
		SET theme_mode = ?, language = ?, two_fa_enabled = ?, notifications = ?, updated_at = ?
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		settings.ThemeMode,
		settings.Language,
		settings.TwoFAEnabled,
		settings.Notifications,
		time.Now(),
		settings.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user settings not found")
	}

	return nil
}

func (r *MySQLUserSettingsRepository) Delete(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	query := "DELETE FROM user_settings WHERE user_id = ?"

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user settings not found")
	}

	return nil
}
