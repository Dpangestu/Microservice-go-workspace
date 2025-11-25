package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/interfaces/http/response"
	"bkc_microservice/services/user-service/internal/middleware"
	"bkc_microservice/services/user-service/internal/shared"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService services.UserService
	logger      shared.Logger
}

func NewUserHandler(userService services.UserService, logger shared.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// ListUsers godoc
// GET /api/v1/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := 1
	size := 20
	search := ""

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

	if q := r.URL.Query().Get("search"); q != "" {
		search = q
	}

	users, total, err := h.userService.ListUsers(r.Context(), search, page, size)
	if err != nil {
		h.logger.Error("ListUsers", "Failed to list users", err)
		response.InternalServerError(w, err.Error())
		return
	}

	// pagination := &response.Pagination{
	// 	Page:  page,
	// 	Size:  size,
	// 	Total: total,
	// }

	response.OKPaged(w, users, page, size, total)
}

// GetUser godoc
// GET /api/v1/users/:id
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, "User not found")
			return
		}
		h.logger.Error("GetUser", "Failed to get user", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, user)
}

// GetCurrentUser godoc
// GET /me
// Returns current logged-in user data from JWT/gateway claims
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// 1. Extract JWT claims from context (set by gateway middleware)
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims.UserID == "" {
		response.Unauthorized(w, "missing user claims")
		return
	}

	// 2. TRY CACHE FIRST (95% of requests should hit here)
	cachedBundle, err := h.userService.GetCachedUserBundle(r.Context(), claims.UserID)
	if cachedBundle != nil && err == nil {
		h.logger.Info("GetCurrentUser", fmt.Sprintf("Cache HIT for user %s", claims.UserID))

		// Apply field masking based on scope
		maskedData := services.MaskUserBundleByScope(cachedBundle, claims.Scope)

		// Log access asynchronously (non-blocking)
		go h.logUserProfileAccess(claims.UserID, claims.ClientID, r)

		response.OK(w, maskedData)
		return
	}

	// 3. CACHE MISS: Query database
	h.logger.Info("GetCurrentUser", fmt.Sprintf("Cache MISS for user %s", claims.UserID))

	user, err := h.userService.GetCurrentUserBundle(r.Context(), claims.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, "User not found")
			return
		}
		h.logger.Error("GetCurrentUser", "Failed to get current user", err)
		response.InternalServerError(w, err.Error())
		return
	}

	// 4. ADD METADATA FROM JWT CLAIMS
	user.Meta.TenantID = claims.TenantID
	user.Meta.ClientID = claims.ClientID
	user.Meta.Scope = claims.Scope

	// 5. CACHE THE RESULT for next 5 minutes
	var userMap map[string]interface{}
	userBytes, _ := json.Marshal(user)
	json.Unmarshal(userBytes, &userMap)
	cacheErr := h.userService.CacheUserBundle(r.Context(), claims.UserID, userMap)
	if cacheErr != nil {
		h.logger.Info("GetCurrentUser", fmt.Sprintf("Cache store error (non-blocking): %v", cacheErr))
	}

	// 6. APPLY FIELD MASKING based on JWT scope
	maskedData := services.MaskUserBundleByScope(userMap, claims.Scope)

	// 7. LOG ACCESS ASYNCHRONOUSLY (non-blocking, won't slow down response)
	go h.logUserProfileAccess(claims.UserID, claims.ClientID, r)

	// 8. RETURN RESPONSE
	response.OK(w, maskedData)
}

// CreateUser godoc
// POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req services.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validation
	if req.Username == "" {
		response.BadRequest(w, "Username is required")
		return
	}
	if len(req.Username) < 3 {
		response.BadRequest(w, "Username must be at least 3 characters")
		return
	}

	if req.Email == "" {
		response.BadRequest(w, "Email is required")
		return
	}

	if req.Password == "" {
		response.BadRequest(w, "Password is required")
		return
	}
	if len(req.Password) < 8 {
		response.BadRequest(w, "Password must be at least 8 characters")
		return
	}

	if req.RoleID <= 0 {
		response.BadRequest(w, "Valid Role ID is required")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		if err.Error() == "role not found" {
			response.BadRequest(w, "Role not found")
			return
		}
		h.logger.Error("CreateUser", "Failed to create user", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, user)
}

// UpdateUser godoc
// PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	var req services.UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// At least one field should be provided
	if req.Username == nil && req.Email == nil && req.RoleID == nil && req.IsActive == nil {
		response.BadRequest(w, "At least one field must be provided for update")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, "User not found")
			return
		}
		if err.Error() == "role not found" {
			response.BadRequest(w, "Role not found")
			return
		}
		h.logger.Error("UpdateUser", "Failed to update user", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, user)
}

// DeleteUser godoc
// DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, "User not found")
			return
		}
		h.logger.Error("DeleteUser", "Failed to delete user", err)
		response.InternalServerError(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *UserHandler) logUserProfileAccess(userID, clientID string, r *http.Request) {
	defer func() {
		if recover() != nil {
			h.logger.Error("logUserProfileAccess", "panic recovered", nil)
		}
	}()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")

	h.logger.Info("AuditLog",
		fmt.Sprintf("User %s accessed profile (client: %s, ip: %s, ua: %s)",
			userID, clientID, ip, userAgent),
	)
}
