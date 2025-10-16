package repositories

import "bkc_microservice/services/user-service/internal/domain/entities"

type UserRepository interface {
	FindByID(id string) (*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
	Create(user *entities.User) error
	Update(user *entities.User) error
	Delete(id string) error
}

type RoleRepository interface {
	GetAll() ([]*entities.Role, error)
	FindByID(id string) (*entities.Role, error)
}

type PermissionRepository interface {
	GetAll() ([]*entities.Permission, error)
	FindByRoleID(roleID string) ([]*entities.Permission, error)
}
