package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// TransactionUser bundle semua operasi user dalam transaction
type TransactionUser struct {
	db *sql.DB
}

func NewTransactionUser(db *sql.DB) *TransactionUser {
	return &TransactionUser{db: db}
}

// CreateUserWithRole creates user dan assign role dalam satu transaction
func (tu *TransactionUser) CreateUserWithRole(ctx context.Context, user *entities.User, roleID int) error {
	tx, err := tu.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Defer rollback jika ada error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Create user
	query := `
		INSERT INTO users (id, username, email, password_hash, role_id, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, true, NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, roleID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// FIX: Capture 2 return values dari RowsAffected()
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("no rows inserted for user")
	}

	// 2. Log activity (opsional)
	activityQuery := `
		INSERT INTO user_activities (id, user_id, action, resource, created_at)
		VALUES (?, ?, 'CREATE', 'user', NOW())
	`
	_, err = tx.ExecContext(ctx, activityQuery, user.ID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	// Commit jika semua sukses
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateUserWithRoleChange updates user dan ganti role dalam satu transaction
func (tu *TransactionUser) UpdateUserWithRoleChange(
	ctx context.Context,
	userID string,
	username, email string,
	oldRoleID, newRoleID int,
) error {
	tx, err := tu.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Verify role exists (sebelum update)
	var roleExists bool
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) > 0 FROM roles WHERE id = ?", newRoleID).
		Scan(&roleExists)
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}
	if !roleExists {
		return fmt.Errorf("role %d not found", newRoleID)
	}

	// 2. Update user
	query := `
		UPDATE users
		SET username = ?, email = ?, role_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	result, err := tx.ExecContext(ctx, query, username, email, newRoleID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// FIX: Capture 2 return values
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user %s not found or not updated", userID)
	}

	// 3. Log activity
	activityQuery := `
		INSERT INTO user_activities (id, user_id, action, resource, metadata, created_at)
		VALUES (?, ?, 'UPDATE_ROLE', 'user', ?, NOW())
	`
	metadata := fmt.Sprintf(`{"from_role":%d,"to_role":%d}`, oldRoleID, newRoleID)
	_, err = tx.ExecContext(ctx, activityQuery, userID, userID, metadata)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteUserCascade delete user, clear permissions, dan log dalam satu transaction
func (tu *TransactionUser) DeleteUserCascade(ctx context.Context, userID string) error {
	tx, err := tu.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Stronger isolation untuk delete
	})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Soft delete user
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = ?`
	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// FIX: Capture 2 return values
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	// 2. Clear sessions (optional)
	sessionQuery := `DELETE FROM user_sessions WHERE user_id = ?`
	_, err = tx.ExecContext(ctx, sessionQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to clear sessions: %w", err)
	}

	// 3. Log deletion
	activityQuery := `
		INSERT INTO user_activities (id, user_id, action, resource, created_at)
		VALUES (?, ?, 'DELETE', 'user', NOW())
	`
	_, err = tx.ExecContext(ctx, activityQuery, userID, userID)
	if err != nil {
		return fmt.Errorf("failed to log deletion: %w", err)
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
