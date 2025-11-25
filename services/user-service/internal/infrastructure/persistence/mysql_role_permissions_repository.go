package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

type MySQLRolePermissionsRepository struct {
	db *sql.DB
}

func NewMySQLRolePermissionsRepository(db *sql.DB) *MySQLRolePermissionsRepository {
	return &MySQLRolePermissionsRepository{db: db}
}

func (r *MySQLRolePermissionsRepository) GetPermissionsByRoleID(roleID int) ([]*entities.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.resource ASC, p.action ASC
	`

	rows, err := r.db.Query(query, roleID)
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

func (r *MySQLRolePermissionsRepository) GetRolesByPermissionID(permissionID int) ([]*entities.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.level, r.is_active, r.tenant_id, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN role_permissions rp ON r.id = rp.role_id
		WHERE rp.permission_id = ?
		ORDER BY r.level ASC
	`

	rows, err := r.db.Query(query, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*entities.Role, 0)
	for rows.Next() {
		role := &entities.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Level, &role.IsActive, &role.TenantID, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func (r *MySQLRolePermissionsRepository) AssignPermission(ctx context.Context, roleID, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return fmt.Errorf("invalid role or permission ID")
	}

	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	return nil
}

func (r *MySQLRolePermissionsRepository) RevokePermission(ctx context.Context, roleID, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return fmt.Errorf("invalid role or permission ID")
	}

	query := "DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?"

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role-permission mapping not found")
	}

	return nil
}

func (r *MySQLRolePermissionsRepository) RevokeAllPermissions(ctx context.Context, roleID int) error {
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}

	query := "DELETE FROM role_permissions WHERE role_id = ?"

	_, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to revoke all permissions: %w", err)
	}

	return nil
}

func (r *MySQLRolePermissionsRepository) AssignBulk(ctx context.Context, roleID int, permissionIDs []int) error {
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	if len(permissionIDs) == 0 {
		return fmt.Errorf("permission IDs are required")
	}

	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES (?, ?, ?)
	`

	for _, permID := range permissionIDs {
		if permID <= 0 {
			return fmt.Errorf("invalid permission ID: %d", permID)
		}

		_, err := r.db.ExecContext(ctx, query, roleID, permID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to assign permission %d: %w", permID, err)
		}
	}

	return nil
}

func (r *MySQLRolePermissionsRepository) HasPermission(roleID, permissionID int) (bool, error) {
	if roleID <= 0 || permissionID <= 0 {
		return false, fmt.Errorf("invalid role or permission ID")
	}

	query := "SELECT 1 FROM role_permissions WHERE role_id = ? AND permission_id = ?"

	var exists int
	err := r.db.QueryRow(query, roleID, permissionID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return true, nil
}
