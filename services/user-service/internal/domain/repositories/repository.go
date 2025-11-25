package repositories

import (
	"context"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// UserRepository defines user persistence operations
type UserRepository interface {
	ListPaged(ctx context.Context, search string, page, size int) ([]*entities.User, int, error)
	FindByID(id string) (*entities.User, error)
	FindByUsername(username string) (*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
	Create(user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id string) error
	UpdateLoginAttempts(ctx context.Context, userID string, attempts int) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

// RoleRepository defines role persistence operations
type RoleRepository interface {
	GetAllPaged(ctx context.Context, page, size int) ([]*entities.Role, int, error)
	GetAll() ([]*entities.Role, error)
	FindByID(id int) (*entities.Role, error)
	FindByName(name string) (*entities.Role, error)
	Create(ctx context.Context, role *entities.Role) error
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id int) error
}

// PermissionRepository defines permission persistence operations
type PermissionRepository interface {
	GetAllPaged(ctx context.Context, page, size int) ([]*entities.Permission, int, error)
	GetAll() ([]*entities.Permission, error)
	FindByID(ctx context.Context, id int) (*entities.Permission, error)
	FindByName(name string) (*entities.Permission, error)
	FindByResource(ctx context.Context, resource string) ([]*entities.Permission, error)
	Create(ctx context.Context, perm *entities.Permission) error
	Update(ctx context.Context, perm *entities.Permission) error
	Delete(ctx context.Context, id int) error
}

// RolePermissionsRepository defines role-permission relationship operations
type RolePermissionsRepository interface {
	GetPermissionsByRoleID(roleID int) ([]*entities.Permission, error)
	GetRolesByPermissionID(permissionID int) ([]*entities.Role, error)
	AssignPermission(ctx context.Context, roleID, permissionID int) error
	RevokePermission(ctx context.Context, roleID, permissionID int) error
	RevokeAllPermissions(ctx context.Context, roleID int) error
	AssignBulk(ctx context.Context, roleID int, permissionIDs []int) error
	HasPermission(roleID, permissionID int) (bool, error)
}

// UserActivityRepository defines user activity audit operations
type UserActivityRepository interface {
	Create(ctx context.Context, activity *entities.UserActivity) error
	GetByUserID(ctx context.Context, userID string, limit int) ([]*entities.UserActivity, error)
	GetByUserIDPaged(ctx context.Context, userID string, page, size int) ([]*entities.UserActivity, int, error)
	DeleteOlderThan(ctx context.Context, days int) error
}

// UserProfileRepository defines user profile operations
type UserProfileRepository interface {
	FindByUserID(userID string) (*entities.UserProfile, error)
	Create(ctx context.Context, profile *entities.UserProfile) error
	Update(ctx context.Context, profile *entities.UserProfile) error
	Delete(ctx context.Context, userID string) error
}

// UserSettingsRepository defines user settings operations
type UserSettingsRepository interface {
	FindByUserID(userID string) (*entities.UserSettings, error)
	Create(ctx context.Context, settings *entities.UserSettings) error
	Update(ctx context.Context, settings *entities.UserSettings) error
	Delete(ctx context.Context, userID string) error
}

// SycCoreUserRepository defines sycrone core user operations
type SycCoreUserRepository interface {
	ListPaged(ctx context.Context, page, size int) ([]*entities.SycCoreUser, int, error)
	FindByID(id string) (*entities.SycCoreUser, error)
	Create(ctx context.Context, sycUser *entities.SycCoreUser) error
	Update(ctx context.Context, sycUser *entities.SycCoreUser) error
	Delete(ctx context.Context, id string) error
}
