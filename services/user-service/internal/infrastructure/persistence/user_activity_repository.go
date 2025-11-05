package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
	"database/sql"
)

type MySQLUserActivityRepository struct {
	DB *sql.DB
}

func NewMySQLUserActivityRepository(db *sql.DB) repositories.UserActivityRepository {
	return &MySQLUserActivityRepository{DB: db}
}

func (r *MySQLUserActivityRepository) Create(activity *entities.UserActivity) error {
	_, err := r.DB.Exec(`INSERT INTO user_activities (id, user_id, action, description, ip, user_agent, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		activity.ID, activity.UserID, activity.Action, activity.Description, activity.IPAddress, activity.UserAgent)
	return err
}

func (r *MySQLUserActivityRepository) GetByUserID(userID string) ([]*entities.UserActivity, error) {
	rows, err := r.DB.Query("SELECT id, user_id, action, description, ip, user_agent, created_at FROM user_activities WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*entities.UserActivity
	for rows.Next() {
		activity := &entities.UserActivity{}
		if err := rows.Scan(&activity.ID, &activity.UserID, &activity.Action, &activity.Description, &activity.IPAddress, &activity.UserAgent, &activity.CreatedAt); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}
