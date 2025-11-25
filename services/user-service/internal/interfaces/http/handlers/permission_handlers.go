package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/interfaces/http/response"
	"bkc_microservice/services/user-service/internal/shared"

	"github.com/gorilla/mux"
)

type PermissionHandler struct {
	permService services.PermissionService
	logger      shared.Logger
}

func NewPermissionHandler(permService services.PermissionService, logger shared.Logger) *PermissionHandler {
	return &PermissionHandler{
		permService: permService,
		logger:      logger,
	}
}

// ListPermissions godoc
// GET /api/v1/permissions
func (h *PermissionHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	page := 1
	size := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}

	if s := r.URL.Query().Get("size"); s != "" {
		if si, err := strconv.Atoi(s); err == nil && si > 0 {
			if si > 100 {
				si = 100
			}
			size = si
		}
	}

	permissions, total, err := h.permService.ListPermissions(r.Context(), page, size)
	if err != nil {
		h.logger.Error("ListPermissions", "Failed to list permissions", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OKPaged(w, permissions, page, size, total)
}

// GetPermission godoc
// GET /api/v1/permissions/:id
func (h *PermissionHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	log.Printf("log param hendler get permission: %v", idStr)
	if idStr == "" {
		response.BadRequest(w, "Permission ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Permission ID")
		return
	}

	permission, err := h.permService.GetPermissionByID(r.Context(), id)
	if err != nil {
		if err.Error() == "permission not found" {
			response.NotFound(w, "Permission not found")
			return
		}
		h.logger.Error("GetPermission", "Failed to get permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, permission)
}

// GetPermissionsByResource godoc
// GET /api/v1/permissions/resource/:resource
func (h *PermissionHandler) GetPermissionsByResource(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	if resource == "" {
		response.BadRequest(w, "Resource is required")
		return
	}

	permissions, err := h.permService.GetPermissionsByResource(r.Context(), resource)
	if err != nil {
		h.logger.Error("GetPermissionsByResource", "Failed to get permissions", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, permissions)
}

// CreatePermission godoc
// POST /api/v1/permissions
func (h *PermissionHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req services.CreatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Name == "" {
		response.BadRequest(w, "Permission name is required")
		return
	}
	if len(req.Name) < 3 {
		response.BadRequest(w, "Permission name must be at least 3 characters")
		return
	}

	if req.Resource == "" {
		response.BadRequest(w, "Resource is required")
		return
	}
	if len(req.Resource) < 2 {
		response.BadRequest(w, "Resource must be at least 2 characters")
		return
	}

	if req.Action == "" {
		response.BadRequest(w, "Action is required")
		return
	}
	if len(req.Action) < 2 {
		response.BadRequest(w, "Action must be at least 2 characters")
		return
	}

	permission, err := h.permService.CreatePermission(r.Context(), &req)
	if err != nil {
		h.logger.Error("CreatePermission", "Failed to create permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, permission)
}

// UpdatePermission godoc
// PUT /api/v1/permissions/:id
func (h *PermissionHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	log.Printf("log param hendler update permission: %v", idStr)
	if idStr == "" {
		response.BadRequest(w, "Permission ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Permission ID")
		return
	}

	var req services.UpdatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	permission, err := h.permService.UpdatePermission(r.Context(), id, &req)
	if err != nil {
		if err.Error() == "permission not found" {
			response.NotFound(w, "Permission not found")
			return
		}
		h.logger.Error("UpdatePermission", "Failed to update permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, permission)
}

// DeletePermission godoc
// DELETE /api/v1/permissions/:id
func (h *PermissionHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	log.Printf("log param hendler delete permission: %v", idStr)
	if idStr == "" {
		response.BadRequest(w, "Permission ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Permission ID")
		return
	}

	if err := h.permService.DeletePermission(r.Context(), id); err != nil {
		if err.Error() == "permission not found" {
			response.NotFound(w, "Permission not found")
			return
		}
		h.logger.Error("DeletePermission", "Failed to delete permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// AssignPermissionToRole godoc
// POST /api/v1/roles/:roleId/permissions/:permissionId
func (h *PermissionHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleStr := r.PathValue("roleId")
	permStr := r.PathValue("permissionId")

	if roleStr == "" || permStr == "" {
		response.BadRequest(w, "Role ID and Permission ID are required")
		return
	}

	roleID, err := strconv.Atoi(roleStr)
	if err != nil || roleID <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	permID, err := strconv.Atoi(permStr)
	if err != nil || permID <= 0 {
		response.BadRequest(w, "Invalid Permission ID")
		return
	}

	if err := h.permService.AssignPermissionToRole(r.Context(), roleID, permID); err != nil {
		if err.Error() == "permission already assigned to this role" {
			response.Conflict(w, "Permission already assigned to this role")
			return
		}
		h.logger.Error("AssignPermissionToRole", "Failed to assign permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "Permission assigned successfully"})
}

// RevokePermissionFromRole godoc
// DELETE /api/v1/roles/:roleId/permissions/:permissionId
func (h *PermissionHandler) RevokePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleStr := r.PathValue("roleId")
	permStr := r.PathValue("permissionId")

	if roleStr == "" || permStr == "" {
		response.BadRequest(w, "Role ID and Permission ID are required")
		return
	}

	roleID, err := strconv.Atoi(roleStr)
	if err != nil || roleID <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	permID, err := strconv.Atoi(permStr)
	if err != nil || permID <= 0 {
		response.BadRequest(w, "Invalid Permission ID")
		return
	}

	if err := h.permService.RevokePermissionFromRole(r.Context(), roleID, permID); err != nil {
		h.logger.Error("RevokePermissionFromRole", "Failed to revoke permission", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// AssignBulkPermissions godoc
// POST /api/v1/roles/:roleId/permissions/bulk
func (h *PermissionHandler) AssignBulkPermissions(w http.ResponseWriter, r *http.Request) {
	roleStr := r.PathValue("roleId")
	if roleStr == "" {
		response.BadRequest(w, "Role ID is required")
		return
	}

	roleID, err := strconv.Atoi(roleStr)
	if err != nil || roleID <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	var req services.AssignPermissionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.PermissionIDs) == 0 {
		response.BadRequest(w, "At least one permission ID is required")
		return
	}

	if err := h.permService.AssignBulkPermissions(r.Context(), roleID, req.PermissionIDs); err != nil {
		h.logger.Error("AssignBulkPermissions", "Failed to assign permissions", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, map[string]string{"message": "Permissions assigned successfully"})
}

// GetRolePermissions godoc
// GET /api/v1/roles/:roleId/permissions
func (h *PermissionHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleStr := r.PathValue("roleId")
	if roleStr == "" {
		response.BadRequest(w, "Role ID is required")
		return
	}

	roleID, err := strconv.Atoi(roleStr)
	if err != nil || roleID <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	permissions, err := h.permService.GetPermissionsByRoleID(r.Context(), roleID)
	if err != nil {
		h.logger.Error("GetRolePermissions", "Failed to get role permissions", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, permissions)
}
