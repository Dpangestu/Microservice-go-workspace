package services

import (
	"context"
	"fmt"

	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
)

type PermissionService interface {
	ListPermissions(ctx context.Context, page, size int) ([]*PermissionResponse, int, error)
	GetPermissionByID(ctx context.Context, id int) (*PermissionResponse, error)
	GetPermissionsByResource(ctx context.Context, resource string) ([]*PermissionResponse, error)
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error)
	UpdatePermission(ctx context.Context, id int, req *UpdatePermissionRequest) (*PermissionResponse, error)
	DeletePermission(ctx context.Context, id int) error
	AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error
	RevokePermissionFromRole(ctx context.Context, roleID int, permissionID int) error
	AssignBulkPermissions(ctx context.Context, roleID int, permissionIDs []int) error
	GetPermissionsByRoleID(ctx context.Context, roleID int) ([]*PermissionResponse, error)
}

type permissionServiceImpl struct {
	permRepo repositories.PermissionRepository
	rpRepo   repositories.RolePermissionsRepository
}

func NewPermissionService(
	permRepo repositories.PermissionRepository,
	rpRepo repositories.RolePermissionsRepository,
) PermissionService {
	return &permissionServiceImpl{
		permRepo: permRepo,
		rpRepo:   rpRepo,
	}
}

func (s *permissionServiceImpl) ListPermissions(ctx context.Context, page, size int) ([]*PermissionResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	permissions, total, err := s.permRepo.GetAllPaged(ctx, page, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	responses := make([]*PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = s.entityToResponse(p)
	}

	return responses, total, nil
}

func (s *permissionServiceImpl) GetPermissionByID(ctx context.Context, id int) (*PermissionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	permission, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return s.entityToResponse(permission), nil
}

func (s *permissionServiceImpl) GetPermissionsByResource(ctx context.Context, resource string) ([]*PermissionResponse, error) {
	if resource == "" {
		return nil, fmt.Errorf("resource is required")
	}

	permissions, err := s.permRepo.FindByResource(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by resource: %w", err)
	}

	responses := make([]*PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = s.entityToResponse(p)
	}

	return responses, nil
}

func (s *permissionServiceImpl) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create permission request is required")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("permission name is required")
	}
	if len(req.Name) < 3 {
		return nil, fmt.Errorf("permission name must be at least 3 characters")
	}

	if req.Resource == "" {
		return nil, fmt.Errorf("resource is required")
	}
	if len(req.Resource) < 2 {
		return nil, fmt.Errorf("resource must be at least 2 characters")
	}

	if req.Action == "" {
		return nil, fmt.Errorf("action is required")
	}
	if len(req.Action) < 2 {
		return nil, fmt.Errorf("action must be at least 2 characters")
	}

	permission := &entities.Permission{
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := s.permRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return s.entityToResponse(permission), nil
}

func (s *permissionServiceImpl) UpdatePermission(ctx context.Context, id int, req *UpdatePermissionRequest) (*PermissionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	if req == nil {
		return nil, fmt.Errorf("update request is required")
	}

	permission, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Apply updates if provided
	if req.Name != nil && *req.Name != "" {
		if len(*req.Name) < 3 {
			return nil, fmt.Errorf("permission name must be at least 3 characters")
		}
		permission.Name = *req.Name
	}

	if req.Resource != nil && *req.Resource != "" {
		if len(*req.Resource) < 2 {
			return nil, fmt.Errorf("resource must be at least 2 characters")
		}
		permission.Resource = *req.Resource
	}

	if req.Action != nil && *req.Action != "" {
		if len(*req.Action) < 2 {
			return nil, fmt.Errorf("action must be at least 2 characters")
		}
		permission.Action = *req.Action
	}

	if req.Description != nil {
		permission.Description = *req.Description
	}

	if err := s.permRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return s.entityToResponse(permission), nil
}

func (s *permissionServiceImpl) DeletePermission(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid permission ID")
	}

	// Check if permission exists
	_, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return s.permRepo.Delete(ctx, id)
}

func (s *permissionServiceImpl) AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error {
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	if permissionID <= 0 {
		return fmt.Errorf("invalid permission ID")
	}

	// Verify permission exists
	_, err := s.permRepo.FindByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if already assigned
	hasPermission, err := s.rpRepo.HasPermission(roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if hasPermission {
		return fmt.Errorf("permission already assigned to this role")
	}

	return s.rpRepo.AssignPermission(ctx, roleID, permissionID)
}

func (s *permissionServiceImpl) RevokePermissionFromRole(ctx context.Context, roleID int, permissionID int) error {
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	if permissionID <= 0 {
		return fmt.Errorf("invalid permission ID")
	}

	return s.rpRepo.RevokePermission(ctx, roleID, permissionID)
}

func (s *permissionServiceImpl) AssignBulkPermissions(ctx context.Context, roleID int, permissionIDs []int) error {
	if roleID <= 0 {
		return fmt.Errorf("invalid role ID")
	}
	if len(permissionIDs) == 0 {
		return fmt.Errorf("permission IDs are required")
	}

	// Verify all permissions exist
	for _, permID := range permissionIDs {
		if permID <= 0 {
			return fmt.Errorf("invalid permission ID: %d", permID)
		}
		_, err := s.permRepo.FindByID(ctx, permID)
		if err != nil {
			return fmt.Errorf("permission %d not found: %w", permID, err)
		}
	}

	// Revoke all existing permissions first
	if err := s.rpRepo.RevokeAllPermissions(ctx, roleID); err != nil {
		return fmt.Errorf("failed to revoke existing permissions: %w", err)
	}

	// Assign bulk permissions
	return s.rpRepo.AssignBulk(ctx, roleID, permissionIDs)
}

func (s *permissionServiceImpl) GetPermissionsByRoleID(ctx context.Context, roleID int) ([]*PermissionResponse, error) {
	if roleID <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	permissions, err := s.rpRepo.GetPermissionsByRoleID(roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by role: %w", err)
	}

	responses := make([]*PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = s.entityToResponse(p)
	}

	return responses, nil
}

func (s *permissionServiceImpl) entityToResponse(p *entities.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
