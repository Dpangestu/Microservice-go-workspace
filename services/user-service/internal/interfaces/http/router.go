package http

import (
	"net/http"
	"time"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/interfaces/http/handlers"
	"bkc_microservice/services/user-service/internal/middleware"
	"bkc_microservice/services/user-service/internal/shared"
	shmiddleware "bkc_microservice/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewRouter(
	userService services.UserService,
	roleService services.RoleService,
	permService services.PermissionService,
	logger shared.Logger,
	rdb *redis.Client,
) http.Handler {
	r := mux.NewRouter()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	roleHandler := handlers.NewRoleHandler(roleService, logger)
	permissionHandler := handlers.NewPermissionHandler(permService, logger)

	// ==================== HEALTH CHECK (NO AUTH) ====================
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	// ==================== AUTHENTICATED ROUTES ====================
	authenticatedRouter := r.PathPrefix("/").Subrouter()
	authenticatedRouter.Use(middleware.InjectClaimsFromGateway)

	// GET /me dengan rate limiting (60 req/min per user)
	authenticatedRouter.Handle("/me",
		shmiddleware.RateLimitUserMeEndpoint(rdb, 60, 1*time.Minute)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.GetCurrentUser(w, r)
			}),
		),
	).Methods(http.MethodGet)

	// ==================== API V1 ROUTES ====================
	apiRouter := r.PathPrefix("/api/v1").Subrouter()

	// ==================== USERS ROUTES ====================
	apiRouter.HandleFunc("/users", userHandler.ListUsers).Methods(http.MethodGet)
	apiRouter.HandleFunc("/users", userHandler.CreateUser).Methods(http.MethodPost)
	apiRouter.HandleFunc("/users/{id}", userHandler.GetUser).Methods(http.MethodGet)
	apiRouter.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods(http.MethodPut)
	apiRouter.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods(http.MethodDelete)

	// ==================== ROLES ROUTES ====================
	apiRouter.HandleFunc("/roles", roleHandler.ListRoles).Methods(http.MethodGet)
	apiRouter.HandleFunc("/roles", roleHandler.CreateRole).Methods(http.MethodPost)
	apiRouter.HandleFunc("/roles/{id}", roleHandler.GetRole).Methods(http.MethodGet)
	apiRouter.HandleFunc("/roles/{id}", roleHandler.UpdateRole).Methods(http.MethodPut)
	apiRouter.HandleFunc("/roles/{id}", roleHandler.DeleteRole).Methods(http.MethodDelete)

	// ==================== PERMISSIONS ROUTES ====================
	apiRouter.HandleFunc("/permissions", permissionHandler.ListPermissions).Methods(http.MethodGet)
	apiRouter.HandleFunc("/permissions", permissionHandler.CreatePermission).Methods(http.MethodPost)
	apiRouter.HandleFunc("/permissions/{id}", permissionHandler.GetPermission).Methods(http.MethodGet)
	apiRouter.HandleFunc("/permissions/{id}", permissionHandler.UpdatePermission).Methods(http.MethodPut)
	apiRouter.HandleFunc("/permissions/{id}", permissionHandler.DeletePermission).Methods(http.MethodDelete)
	apiRouter.HandleFunc("/permissions/resource/{resource}", permissionHandler.GetPermissionsByResource).Methods(http.MethodGet)

	// ==================== ROLE-PERMISSIONS ROUTES ====================
	apiRouter.HandleFunc("/roles/{roleId}/permissions", permissionHandler.GetRolePermissions).Methods(http.MethodGet)
	apiRouter.HandleFunc("/roles/{roleId}/permissions/{permissionId}", permissionHandler.AssignPermissionToRole).Methods(http.MethodPost)
	apiRouter.HandleFunc("/roles/{roleId}/permissions/{permissionId}", permissionHandler.RevokePermissionFromRole).Methods(http.MethodDelete)
	apiRouter.HandleFunc("/roles/{roleId}/permissions/bulk", permissionHandler.AssignBulkPermissions).Methods(http.MethodPost)

	return r
}
