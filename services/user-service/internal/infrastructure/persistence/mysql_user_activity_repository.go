package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

type MySQLUserActivityRepository struct {
	db *sql.DB
}

func NewMySQLUserActivityRepository(db *sql.DB) *MySQLUserActivityRepository {
	return &MySQLUserActivityRepository{db: db}
}

func (r *MySQLUserActivityRepository) Create(ctx context.Context, activity *entities.UserActivity) error {
	if activity == nil {
		return fmt.Errorf("activity is required")
	}

	query := `
		INSERT INTO user_activities (user_id, action, resource, details, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		activity.UserID,
		activity.Action,
		activity.Resource,
		activity.Description,
		activity.IPAddress,
		activity.UserAgent,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create user activity: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	activity.ID = int(id)
	return nil
}

func (r *MySQLUserActivityRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*entities.UserActivity, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, user_id, action, resource, details, ip_address, user_agent, created_at
		FROM user_activities
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query user activities: %w", err)
	}
	defer rows.Close()

	activities := make([]*entities.UserActivity, 0)
	for rows.Next() {
		activity := &entities.UserActivity{}
		if err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Action,
			&activity.Resource,
			&activity.Description,
			&activity.IPAddress,
			&activity.UserAgent,
			&activity.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, rows.Err()
}

func (r *MySQLUserActivityRepository) GetByUserIDPaged(ctx context.Context, userID string, page, size int) ([]*entities.UserActivity, int, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size

	// Get total count
	countQuery := "SELECT COUNT(*) FROM user_activities WHERE user_id = ?"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	query := `
		SELECT id, user_id, action, resource, details, ip_address, user_agent, created_at
		FROM user_activities
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user activities: %w", err)
	}
	defer rows.Close()

	activities := make([]*entities.UserActivity, 0)
	for rows.Next() {
		activity := &entities.UserActivity{}
		if err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Action,
			&activity.Resource,
			&activity.Description,
			&activity.IPAddress,
			&activity.UserAgent,
			&activity.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, activity)
	}

	return activities, total, rows.Err()
}

func (r *MySQLUserActivityRepository) DeleteOlderThan(ctx context.Context, days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be greater than 0")
	}

	query := "DELETE FROM user_activities WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)"

	result, err := r.db.ExecContext(ctx, query, days)
	if err != nil {
		return fmt.Errorf("failed to delete old activities: %w", err)
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	fmt.Printf("Deleted %d old activities\n", rowsDeleted)
	return nil
}
