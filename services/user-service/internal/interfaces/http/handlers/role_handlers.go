package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/interfaces/http/response"
	"bkc_microservice/services/user-service/internal/shared"

	"github.com/gorilla/mux"
)

type RoleHandler struct {
	roleService services.RoleService
	logger      shared.Logger
}

func NewRoleHandler(roleService services.RoleService, logger shared.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		logger:      logger,
	}
}

// ListRoles godoc
// GET /api/v1/roles
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
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

	roles, total, err := h.roleService.ListRoles(r.Context(), page, size)
	if err != nil {
		h.logger.Error("ListRoles", "Failed to list roles", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OKPaged(w, roles, page, size, total)
}

// GetRole godoc
// GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		response.BadRequest(w, "Role ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	role, err := h.roleService.GetRoleByID(r.Context(), id)
	if err != nil {
		if err.Error() == "role not found" {
			response.NotFound(w, "Role not found")
			return
		}
		h.logger.Error("GetRole", "Failed to get role", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, role)
}

// CreateRole godoc
// POST /api/v1/roles
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req services.CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Name == "" {
		response.BadRequest(w, "Role name is required")
		return
	}
	if len(req.Name) < 2 {
		response.BadRequest(w, "Role name must be at least 2 characters")
		return
	}

	if req.Level <= 0 {
		response.BadRequest(w, "Role level must be greater than 0")
		return
	}

	role, err := h.roleService.CreateRole(r.Context(), &req)
	if err != nil {
		h.logger.Error("CreateRole", "Failed to create role", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, role)
}

// UpdateRole godoc
// PUT /api/v1/roles/:id
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		response.BadRequest(w, "Role ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	var req services.UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// At least one field should be provided
	if req.Name == nil && req.Description == nil && req.Level == nil && req.IsActive == nil {
		response.BadRequest(w, "At least one field must be provided for update")
		return
	}

	role, err := h.roleService.UpdateRole(r.Context(), id, &req)
	if err != nil {
		if err.Error() == "role not found" {
			response.NotFound(w, "Role not found")
			return
		}
		h.logger.Error("UpdateRole", "Failed to update role", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, role)
}

// DeleteRole godoc
// DELETE /api/v1/roles/:id
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		response.BadRequest(w, "Role ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response.BadRequest(w, "Invalid Role ID")
		return
	}

	if err := h.roleService.DeleteRole(r.Context(), id); err != nil {
		if err.Error() == "role not found" {
			response.NotFound(w, "Role not found")
			return
		}
		h.logger.Error("DeleteRole", "Failed to delete role", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.NoContent(w)
}
