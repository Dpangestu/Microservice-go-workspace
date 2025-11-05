package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
	"database/sql"
)

type MySQLPermissionRepository struct {
	DB *sql.DB
}

func NewMySQLPermissionRepository(db *sql.DB) repositories.PermissionRepository {
	return &MySQLPermissionRepository{DB: db}
}

func (r *MySQLPermissionRepository) GetAll() ([]*entities.Permission, error) {
	rows, err := r.DB.Query("SELECT id, name, resource, action, description, created_at, updated_at FROM permissions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		permission := &entities.Permission{}
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func (r *MySQLPermissionRepository) FindByRoleID(roleID string) ([]*entities.Permission, error) {
	rows, err := r.DB.Query("SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id WHERE rp.role_id = ?", roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		permission := &entities.Permission{}
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Resource, &permission.Action, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
