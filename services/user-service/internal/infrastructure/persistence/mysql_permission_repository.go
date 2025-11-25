package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// MySQLPermissionRepository implements PermissionRepository interface
type MySQLPermissionRepository struct {
	db *sql.DB
}

func NewMySQLPermissionRepository(db *sql.DB) *MySQLPermissionRepository {
	return &MySQLPermissionRepository{db: db}
}

func (r *MySQLPermissionRepository) GetAllPaged(ctx context.Context, page, size int) ([]*entities.Permission, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	offset := (page - 1) * size

	countQuery := "SELECT COUNT(*) FROM permissions"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		ORDER BY id ASC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := make([]*entities.Permission, 0)
	for rows.Next() {
		perm := &entities.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, total, rows.Err()
}

func (r *MySQLPermissionRepository) GetAll() ([]*entities.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := make([]*entities.Permission, 0)
	for rows.Next() {
		perm := &entities.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

func (r *MySQLPermissionRepository) FindByID(ctx context.Context, id int) (*entities.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE id = ?
	`

	perm := &entities.Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

	return perm, nil
}

func (r *MySQLPermissionRepository) FindByName(name string) (*entities.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE name = ?
	`

	perm := &entities.Permission{}
	err := r.db.QueryRow(query, name).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

	return perm, nil
}

func (r *MySQLPermissionRepository) FindByResource(ctx context.Context, resource string) ([]*entities.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE resource = ?
		ORDER BY action ASC
	`

	rows, err := r.db.QueryContext(ctx, query, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	permissions := make([]*entities.Permission, 0)
	for rows.Next() {
		perm := &entities.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}

func (r *MySQLPermissionRepository) Create(ctx context.Context, perm *entities.Permission) error {
	if perm == nil {
		return fmt.Errorf("permission is required")
	}

	query := `
		INSERT INTO permissions (name, resource, action, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, perm.Name, perm.Resource, perm.Action, perm.Description, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	perm.ID = int(id)
	return nil
}

func (r *MySQLPermissionRepository) Update(ctx context.Context, perm *entities.Permission) error {
	if perm == nil {
		return fmt.Errorf("permission is required")
	}

	query := `
		UPDATE permissions
		SET name = ?, resource = ?, action = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, perm.Name, perm.Resource, perm.Action, perm.Description, time.Now(), perm.ID)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

func (r *MySQLPermissionRepository) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid permission ID")
	}

	query := "DELETE FROM permissions WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}
