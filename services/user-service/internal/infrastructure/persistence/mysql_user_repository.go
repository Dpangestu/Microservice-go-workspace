package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"

	"github.com/google/uuid"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) ListPaged(ctx context.Context, search string, page, size int) ([]*entities.User, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users WHERE is_active = true"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build query with search
	query := `
		SELECT id, username, email, role_id, is_active, is_locked, failed_login_attempts, 
		       last_login, created_at, updated_at
		FROM users
		WHERE is_active = true
	`

	args := []interface{}{}
	if search != "" {
		query += ` AND (username LIKE ? OR email LIKE ?)`
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, size, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.RoleID,
			&user.IsActive,
			&user.IsLocked,
			&user.FailedLoginAttempts,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

func (r *MySQLUserRepository) FindByID(id string) (*entities.User, error) {
	query := `
		SELECT id, username, email, role_id, is_active, is_locked, failed_login_attempts,
		       last_login, created_at, updated_at
		FROM users
		WHERE id = ? AND is_active = true
	`

	user := &entities.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.RoleID,
		&user.IsActive,
		&user.IsLocked,
		&user.FailedLoginAttempts,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (r *MySQLUserRepository) FindByUsername(username string) (*entities.User, error) {
	query := `
		SELECT id, username, email, role_id, is_active, is_locked, failed_login_attempts,
		       last_login, created_at, updated_at
		FROM users
		WHERE username = ? AND is_active = true
	`

	user := &entities.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.RoleID,
		&user.IsActive,
		&user.IsLocked,
		&user.FailedLoginAttempts,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (r *MySQLUserRepository) FindByEmail(email string) (*entities.User, error) {
	query := `
		SELECT id, username, email, role_id, is_active, is_locked, failed_login_attempts,
		       last_login, created_at, updated_at
		FROM users
		WHERE email = ? AND is_active = true
	`

	user := &entities.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.RoleID,
		&user.IsActive,
		&user.IsLocked,
		&user.FailedLoginAttempts,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (r *MySQLUserRepository) Create(user *entities.User) error {
	if user == nil {
		return fmt.Errorf("user is required")
	}

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	query := `
        INSERT INTO users (id, username, email, password_hash, role_id, is_active, 
                           is_locked, failed_login_attempts, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.RoleID,
		user.IsActive,
		user.IsLocked,
		user.FailedLoginAttempts,
		time.Now(),
	)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *entities.User) error {
	if user == nil {
		return fmt.Errorf("user is required")
	}

	query := `
		UPDATE users
		SET username = ?, email = ?, role_id = ?, is_active = ?, is_locked = ?, 
		    failed_login_attempts = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.RoleID,
		user.IsActive,
		user.IsLocked,
		user.FailedLoginAttempts,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *MySQLUserRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	query := "UPDATE users SET is_active = false, updated_at = NOW() WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *MySQLUserRepository) UpdateLoginAttempts(ctx context.Context, userID string, attempts int) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	query := "UPDATE users SET failed_login_attempts = ? WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, attempts, userID)
	if err != nil {
		return fmt.Errorf("failed to update login attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *MySQLUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	query := "UPDATE users SET last_login = ?, failed_login_attempts = 0 WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
