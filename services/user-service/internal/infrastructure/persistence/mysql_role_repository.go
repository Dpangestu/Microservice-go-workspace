package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

type MySQLRoleRepository struct {
	db *sql.DB
}

func NewMySQLRoleRepository(db *sql.DB) *MySQLRoleRepository {
	return &MySQLRoleRepository{db: db}
}

func (r *MySQLRoleRepository) GetAllPaged(ctx context.Context, page, size int) ([]*entities.Role, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size

	// Get total count
	countQuery := "SELECT COUNT(*) FROM roles WHERE is_active = true"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	query := `
		SELECT id, name, description, level, is_active, tenant_id, created_at, updated_at
		FROM roles
		WHERE is_active = true
		ORDER BY level ASC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*entities.Role, 0)
	for rows.Next() {
		role := &entities.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.Level,
			&role.IsActive,
			&role.TenantID,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, total, rows.Err()
}

func (r *MySQLRoleRepository) GetAll() ([]*entities.Role, error) {
	query := `
		SELECT id, name, description, level, is_active, tenant_id, created_at, updated_at
		FROM roles
		WHERE is_active = true
		ORDER BY id ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*entities.Role, 0)
	for rows.Next() {
		role := &entities.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.Level,
			&role.IsActive,
			&role.TenantID,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *MySQLRoleRepository) FindByID(id int) (*entities.Role, error) {
	query := `
		SELECT id, name, description, level, is_active, tenant_id, created_at, updated_at
		FROM roles
		WHERE id = ? AND is_active = true
	`

	role := &entities.Role{}
	err := r.db.QueryRow(query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
		&role.IsActive,
		&role.TenantID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return role, nil
}

func (r *MySQLRoleRepository) FindByName(name string) (*entities.Role, error) {
	query := `
		SELECT id, name, description, level, is_active, tenant_id, created_at, updated_at
		FROM roles
		WHERE name = ? AND is_active = true
	`

	role := &entities.Role{}
	err := r.db.QueryRow(query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
		&role.IsActive,
		&role.TenantID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return role, nil
}

func (r *MySQLRoleRepository) Create(ctx context.Context, role *entities.Role) error {
	if role == nil {
		return fmt.Errorf("role is required")
	}

	query := `
		INSERT INTO roles (name, description, level, is_active, tenant_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		role.Name,
		role.Description,
		role.Level,
		role.IsActive,
		role.TenantID,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	role.ID = int(id)
	return nil
}

func (r *MySQLRoleRepository) Update(ctx context.Context, role *entities.Role) error {
	if role == nil {
		return fmt.Errorf("role is required")
	}

	query := `
		UPDATE roles
		SET name = ?, description = ?, level = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		role.Name,
		role.Description,
		role.Level,
		role.IsActive,
		time.Now(),
		role.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

func (r *MySQLRoleRepository) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid role ID")
	}

	query := "UPDATE roles SET is_active = false, updated_at = ? WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}
