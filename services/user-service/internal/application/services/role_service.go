package services

import (
	"context"
	"fmt"

	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
)

type RoleService interface {
	ListRoles(ctx context.Context, page, size int) ([]*RoleResponse, int, error)
	GetRoleByID(ctx context.Context, id int) (*RoleResponse, error)
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error)
	UpdateRole(ctx context.Context, id int, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id int) error
}

type roleServiceImpl struct {
	roleRepo repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) RoleService {
	return &roleServiceImpl{
		roleRepo: roleRepo,
	}
}

func (s *roleServiceImpl) ListRoles(ctx context.Context, page, size int) ([]*RoleResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	roles, total, err := s.roleRepo.GetAllPaged(ctx, page, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	responses := make([]*RoleResponse, len(roles))
	for i, r := range roles {
		responses[i] = s.entityToResponse(r)
	}

	return responses, total, nil
}

func (s *roleServiceImpl) GetRoleByID(ctx context.Context, id int) (*RoleResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return s.entityToResponse(role), nil
}

func (s *roleServiceImpl) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create role request is required")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}
	if len(req.Name) < 2 {
		return nil, fmt.Errorf("role name must be at least 2 characters")
	}

	if req.Level <= 0 {
		return nil, fmt.Errorf("role level must be greater than 0")
	}

	role := &entities.Role{
		Name:        req.Name,
		Description: req.Description,
		Level:       req.Level,
		IsActive:    req.IsActive,
		TenantID:    req.TenantID,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return s.entityToResponse(role), nil
}

func (s *roleServiceImpl) UpdateRole(ctx context.Context, id int, req *UpdateRoleRequest) (*RoleResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	if req == nil {
		return nil, fmt.Errorf("update request is required")
	}

	// Check if at least one field is being updated
	if req.Name == nil && req.Description == nil && req.Level == nil && req.IsActive == nil {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Apply updates if provided
	if req.Name != nil && *req.Name != "" {
		if len(*req.Name) < 2 {
			return nil, fmt.Errorf("role name must be at least 2 characters")
		}
		role.Name = *req.Name
	}

	if req.Description != nil {
		role.Description = *req.Description
	}

	if req.Level != nil {
		if *req.Level <= 0 {
			return nil, fmt.Errorf("role level must be greater than 0")
		}
		role.Level = *req.Level
	}

	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return s.entityToResponse(role), nil
}

func (s *roleServiceImpl) DeleteRole(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid role ID")
	}

	// Verify role exists
	_, err := s.roleRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.roleRepo.Delete(ctx, id)
}

func (s *roleServiceImpl) entityToResponse(r *entities.Role) *RoleResponse {
	return &RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Level:       r.Level,
		IsActive:    r.IsActive,
		TenantID:    r.TenantID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
