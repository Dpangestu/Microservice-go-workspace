package repositories

import (
	"bkc_microservice/services/user-service/internal/domain/entities"
)

type UserRepository interface {
	ListAll() ([]*entities.User, error)
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

type RolePermissionsRepository interface {
	GetPermissionsByRoleID(roleID int) ([]*entities.Permission, error)
	GetRoleName(roleID int) (*string, error)
}

type PermissionRepository interface {
	GetAll() ([]*entities.Permission, error)
	FindByRoleID(roleID string) ([]*entities.Permission, error)
}

type UserActivityRepository interface {
	Create(activity *entities.UserActivity) error
	GetByUserID(userID string) ([]*entities.UserActivity, error)
}

type UserProfileRepository interface {
	FindByUserID(userID string) (*entities.UserProfile, error)
	Create(profile *entities.UserProfile) error
	Update(profile *entities.UserProfile) error
}

type UserSettingsRepository interface {
	GetByUserID(userID string) (map[string]string, error)
	Set(userID, key, value string) error
	Delete(userID, key string) error
}
