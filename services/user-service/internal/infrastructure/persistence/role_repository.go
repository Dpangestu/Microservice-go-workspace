package persistence

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
	"database/sql"
)

type MySQLRoleRepository struct {
	DB *sql.DB
}

func NewRoleRepository(db *sql.DB) repositories.RoleRepository {
	return &MySQLRoleRepository{DB: db}
}

func (r *MySQLRoleRepository) GetAll() ([]*entities.Role, error) {
	rows, err := r.DB.Query("SELECT id, tenant_id, name, description, level, is_system, is_active, created_at, updated_at FROM roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role := &entities.Role{}
		if err := rows.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.Level, &role.IsSystem, &role.IsActive, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *MySQLRoleRepository) FindByID(id string) (*entities.Role, error) {
	row := r.DB.QueryRow("SELECT id, tenant_id, name, description, level, is_system, is_active, created_at, updated_at FROM roles WHERE id = ?", id)
	role := &entities.Role{}
	if err := row.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.Level, &role.IsSystem, &role.IsActive, &role.CreatedAt, &role.UpdatedAt); err != nil {
		return nil, err
	}
	return role, nil
}
