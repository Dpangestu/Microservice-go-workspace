package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"database/sql"
	"log"
)

type MySQLRolePermissionsRepository struct {
	DB *sql.DB
}

func NewMySQLRolePermissionsRepository(db *sql.DB) *MySQLRolePermissionsRepository {
	return &MySQLRolePermissionsRepository{DB: db}
}

// GetPermissionsByRoleID mengambil permissions dalam format resource:action
func (r *MySQLRolePermissionsRepository) GetPermissionsByRoleID(roleID int) ([]*entities.Permission, error) {
	rows, err := r.DB.Query(`
        SELECT p.id, p.resource, p.action
        FROM permissions p
        JOIN role_permissions rp ON rp.permission_id = p.id
        WHERE rp.role_id = ?
    `, roleID)
	if err != nil {
		log.Printf("[RolePermissionsRepo] Error fetching permissions for role %s: %v", roleID, err)
		return nil, err
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		var p entities.Permission
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action); err != nil {
			log.Printf("[RolePermissionsRepo] Error scanning permission: %v", err)
			return nil, err
		}
		permissions = append(permissions, &p)
	}

	return permissions, nil
}

// GetRoleName mengambil nama role berdasarkan role_id
func (r *MySQLRolePermissionsRepository) GetRoleName(roleID int) (*string, error) {
	row := r.DB.QueryRow("SELECT name FROM roles WHERE id = ?", roleID)

	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("[RolePermissionsRepo] Error fetching role name for ID %d: %v", roleID, err)
		return nil, err
	}

	return &name, nil
}
