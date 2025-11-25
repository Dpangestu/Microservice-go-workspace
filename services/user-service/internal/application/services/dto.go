package services

import "time"

// ==================== REQUEST DTOs ====================

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   int    `json:"roleId" validate:"required,gt=0"`
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	RoleID   *int    `json:"roleId,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Description string  `json:"description"`
	Level       int     `json:"level" validate:"required,gt=0"`
	IsActive    bool    `json:"isActive"`
	TenantID    *string `json:"tenantId,omitempty"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Level       *int    `json:"level,omitempty,gt=0"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=3"`
	Resource    string `json:"resource" validate:"required,min=2"`
	Action      string `json:"action" validate:"required,min=2"`
	Description string `json:"description"`
}

type UpdatePermissionRequest struct {
	Name        *string `json:"name,omitempty"`
	Resource    *string `json:"resource,omitempty"`
	Action      *string `json:"action,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []int `json:"permissionIds" validate:"required,min=1"`
}

// ==================== RESPONSE DTOs ====================

type UserResponse struct {
	ID                  string     `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	RoleID              int        `json:"roleId"`
	IsActive            bool       `json:"isActive"`
	IsLocked            bool       `json:"isLocked"`
	FailedLoginAttempts int        `json:"failedLoginAttempts"`
	LastLogin           *time.Time `json:"lastLogin,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
}

type RoleResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Level       int        `json:"level"`
	IsActive    bool       `json:"isActive"`
	TenantID    *string    `json:"tenantId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type PermissionResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Resource    string     `json:"resource"`
	Action      string     `json:"action"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type RoleWithPermissionsResponse struct {
	*RoleResponse
	Permissions []*PermissionResponse `json:"permissions,omitempty"`
}

type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
