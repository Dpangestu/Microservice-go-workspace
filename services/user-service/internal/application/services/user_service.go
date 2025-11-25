package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"
	"bkc_microservice/services/user-service/internal/infrastructure/clients"
	"bkc_microservice/services/user-service/internal/infrastructure/persistence"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// MeResponse adalah respons lengkap untuk endpoint /me
type MeResponse struct {
	Data struct {
		ID                  string   `json:"id"`
		Username            string   `json:"username"`
		Email               string   `json:"email"`
		RoleID              *int     `json:"roleId,omitempty"`
		RoleName            *string  `json:"roleName,omitempty"`
		Permissions         []string `json:"permissions,omitempty"`
		IsActive            bool     `json:"isActive"`
		IsLocked            bool     `json:"isLocked"`
		FailedLoginAttempts int      `json:"failedLoginAttempts"`
		LastLogin           *string  `json:"lastLogin,omitempty"`
		CreatedAt           string   `json:"createdAt"`
		UpdatedAt           *string  `json:"updatedAt,omitempty"`
	} `json:"data"`
	Profile *struct {
		FullName    *string         `json:"fullName,omitempty"`
		DisplayName *string         `json:"displayName,omitempty"`
		Phone       *string         `json:"phone,omitempty"`
		AvatarURL   *string         `json:"avatarUrl,omitempty"`
		Locale      *string         `json:"locale,omitempty"`
		Timezone    *string         `json:"timezone,omitempty"`
		Metadata    json.RawMessage `json:"metadata,omitempty"`
	} `json:"profile,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
	CBS      *struct {
		ID            string     `json:"id,omitempty"`
		UserCore      string     `json:"userCore,omitempty"`
		KodeGroup1    string     `json:"kodeGroup1,omitempty"`
		KodePerkiraan string     `json:"kodePerkiraan,omitempty"`
		KodeCabang    *string    `json:"kodeCabang,omitempty"`
		Status        string     `json:"status,omitempty"`
		SyncStatus    string     `json:"syncStatus,omitempty"`
		LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
	} `json:"cbs,omitempty"`
	Meta struct {
		TenantID string `json:"tenantId"`
		ClientID string `json:"clientId"`
		Scope    string `json:"scope"`
	} `json:"meta"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

// ErrorResponse adalah respons error standar
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// UserService interface defines all user operations
type UserService interface {
	ListUsers(ctx context.Context, search string, page, size int) ([]*UserResponse, int, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	GetCurrentUserBundle(ctx context.Context, userID string) (*MeResponse, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	// Cache methods
	CacheUserBundle(ctx context.Context, userID string, data map[string]interface{}) error
	GetCachedUserBundle(ctx context.Context, userID string) (map[string]interface{}, error)
	InvalidateUserCache(ctx context.Context, userID string) error
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	userRepo        repositories.UserRepository
	roleRepo        repositories.RoleRepository
	permissionRepo  repositories.PermissionRepository
	rpRepo          repositories.RolePermissionsRepository
	activityRepo    repositories.UserActivityRepository
	profileRepo     repositories.UserProfileRepository
	settingsRepo    repositories.UserSettingsRepository
	RedisClient     *redis.Client
	syncCBSClient   *clients.SyncCBSClient
	transactionUser *persistence.TransactionUser
}

// NewUserService creates a new user service
// FIX: Parameter list dibenerin + tambah db *sql.DB
func NewUserService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	rpRepo repositories.RolePermissionsRepository,
	activityRepo repositories.UserActivityRepository,
	profileRepo repositories.UserProfileRepository,
	settingsRepo repositories.UserSettingsRepository,
	redisClient *redis.Client,
	syncCBSClient *clients.SyncCBSClient,
	db *sql.DB,
) UserService {
	return &userServiceImpl{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		rpRepo:          rpRepo,
		activityRepo:    activityRepo,
		profileRepo:     profileRepo,
		settingsRepo:    settingsRepo,
		RedisClient:     redisClient,
		syncCBSClient:   syncCBSClient,
		transactionUser: persistence.NewTransactionUser(db), // FIX: call function properly
	}
}

// ListUsers retrieves paginated user list
func (s *userServiceImpl) ListUsers(ctx context.Context, search string, page, size int) ([]*UserResponse, int, error) {
	if page < 1 {
		page = 1
	}

	if size < 1 {
		size = 20
	}

	if size > 100 {
		size = 100
	}

	users, total, err := s.userRepo.ListPaged(ctx, search, page, size)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, u := range users {
		responses[i] = s.entityToResponse(u)
	}

	return responses, total, nil
}

// GetUserByID retrieves a user by ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.entityToResponse(user), nil
}

// GetCurrentUserBundle menggabungkan user, profile, role, permissions, dan settings
// Digunakan untuk endpoint GET /me
func (s *userServiceImpl) GetCurrentUserBundle(ctx context.Context, userID string) (*MeResponse, error) {
	// 1. Ambil user dasar
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		log.Printf("[UserService] Error fetching user %s: %v", userID, err)
		return nil, errors.New("user not found")
	}

	resp := &MeResponse{}

	// 2. Isi data dasar
	resp.Data.ID = user.ID
	resp.Data.Username = user.Username
	resp.Data.Email = user.Email
	resp.Data.IsActive = user.IsActive
	resp.Data.IsLocked = user.IsLocked
	resp.Data.FailedLoginAttempts = user.FailedLoginAttempts
	if user.LastLogin != nil {
		lastLoginStr := user.LastLogin.Format("2006-01-02T15:04:05Z")
		resp.Data.LastLogin = &lastLoginStr
	}
	resp.Data.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z")
	if user.UpdatedAt != nil {
		updatedAtStr := user.UpdatedAt.Format("2006-01-02T15:04:05Z")
		resp.Data.UpdatedAt = &updatedAtStr
	}

	// 3. Ambil role name dan permissions jika ada
	if user.RoleID > 0 {
		resp.Data.RoleID = &user.RoleID

		// Ambil role name
		role, err := s.roleRepo.FindByID(user.RoleID)
		if err == nil && role != nil {
			resp.Data.RoleName = &role.Name
		}

		// Ambil permissions - FIX: convert []*Permission to []string
		permissions, err := s.rpRepo.GetPermissionsByRoleID(user.RoleID)
		if err == nil && len(permissions) > 0 {
			// Convert Permission entities to string array (resource:action format)
			permStrings := make([]string, len(permissions))
			for i, perm := range permissions {
				permStrings[i] = fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
			}
			resp.Data.Permissions = permStrings
		} else {
			resp.Data.Permissions = []string{}
		}
	}

	// 4. Ambil profile
	profile, err := s.profileRepo.FindByUserID(userID)
	if err == nil && profile != nil {
		resp.Profile = &struct {
			FullName    *string         `json:"fullName,omitempty"`
			DisplayName *string         `json:"displayName,omitempty"`
			Phone       *string         `json:"phone,omitempty"`
			AvatarURL   *string         `json:"avatarUrl,omitempty"`
			Locale      *string         `json:"locale,omitempty"`
			Timezone    *string         `json:"timezone,omitempty"`
			Metadata    json.RawMessage `json:"metadata,omitempty"`
		}{
			FullName:    profile.FullName,
			DisplayName: profile.DisplayName,
			Phone:       profile.Phone,
			AvatarURL:   profile.AvatarURL,
			Locale:      profile.Locale,
			Timezone:    profile.Timezone,
			Metadata:    profile.Metadata,
		}
	}

	// 5. Ambil settings
	settings, err := s.settingsRepo.FindByUserID(userID)
	if err == nil && settings != nil {
		resp.Settings = make(map[string]string)
		if settings.ThemeMode != "" {
			resp.Settings["themeMode"] = settings.ThemeMode
		}
		if settings.Language != "" {
			resp.Settings["language"] = settings.Language
		}
	} else {
		resp.Settings = make(map[string]string)
	}

	// 6. Ambil CBS data dari sync-cbs-service
	log.Printf("[UserService] Fetching CBS mapping for user %s", userID)
	cbsData, err := s.syncCBSClient.GetMapping(ctx, userID)
	if err != nil {
		log.Printf("[UserService] Warning: Failed to fetch CBS data: %v", err)
		// Non-blocking: CBS data is optional
	}

	// 7. Add CBS data ke response
	if cbsData != nil {
		resp.CBS = &struct {
			ID            string     `json:"id,omitempty"`
			UserCore      string     `json:"userCore,omitempty"`
			KodeGroup1    string     `json:"kodeGroup1,omitempty"`
			KodePerkiraan string     `json:"kodePerkiraan,omitempty"`
			KodeCabang    *string    `json:"kodeCabang,omitempty"`
			Status        string     `json:"status,omitempty"`
			SyncStatus    string     `json:"syncStatus,omitempty"`
			LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
		}{
			ID:            cbsData.ID,
			UserCore:      cbsData.UserCore,
			KodeGroup1:    cbsData.KodeGroup1,
			KodePerkiraan: cbsData.KodePerkiraan,
			KodeCabang:    cbsData.KodeCabang,
			Status:        cbsData.Status,
			SyncStatus:    cbsData.SyncStatus,
			LastSyncAt:    cbsData.LastSyncAt,
		}
		log.Printf("[UserService] CBS data found for user %s: userCore=%s", userID, cbsData.UserCore)
	}

	// 8. Set links
	resp.Links.Self = "/me"

	return resp, nil
}

// CreateUser creates a new user
func (s *userServiceImpl) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create user request is required")
	}

	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if len(req.Username) < 3 {
		return nil, fmt.Errorf("username must be at least 3 characters")
	}

	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	if req.RoleID <= 0 {
		return nil, fmt.Errorf("valid role ID is required")
	}

	// Verify role exists
	_, err := s.roleRepo.FindByID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entities.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("[UserService] User created: %s", user.ID)

	return s.entityToResponse(user), nil
}

// UpdateUser updates an existing user
func (s *userServiceImpl) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if req == nil {
		return nil, fmt.Errorf("update request is required")
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Apply updates if provided
	if req.Username != nil && *req.Username != "" {
		if len(*req.Username) < 3 {
			return nil, fmt.Errorf("username must be at least 3 characters")
		}
		user.Username = *req.Username
	}

	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}

	if req.RoleID != nil && *req.RoleID > 0 {
		// Verify role exists
		_, err := s.roleRepo.FindByID(*req.RoleID)
		if err != nil {
			return nil, fmt.Errorf("role not found: %w", err)
		}
		user.RoleID = *req.RoleID
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// PENTING: Invalidate cache setelah update
	if err := s.InvalidateUserCache(ctx, id); err != nil {
		log.Printf("⚠️ Warning: cache invalidation failed: %v", err)
	}

	return s.entityToResponse(user), nil
}

// DeleteUser deletes a user (soft delete)
func (s *userServiceImpl) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	// Verify user exists
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// PENTING: Invalidate cache setelah delete
	if err := s.InvalidateUserCache(ctx, id); err != nil {
		log.Printf("⚠️ Warning: cache invalidation failed: %v", err)
	}

	return nil
}

func (s *userServiceImpl) GetProfile(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Get user data
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		log.Printf("[UserService] Error fetching user %s: %v", userID, err)
		return nil, err
	}

	// Get user profile
	profile, err := s.profileRepo.FindByUserID(userID)
	if err != nil {
		log.Printf("[UserService] No profile found for user %s: %v", userID, err)
		profile = nil // Opsional, bisa null
	}

	// Get CBS mapping dari sync-cbs-service
	cbsData, err := s.syncCBSClient.GetMapping(ctx, userID)
	if err != nil {
		log.Printf("[UserService] Error fetching CBS mapping for user %s: %v", userID, err)
		return nil, err
	}

	// Combine data
	response := map[string]interface{}{
		"user":    user,
		"profile": profile,
		"cbs":     cbsData, // nil jika tidak ada
	}

	log.Printf("[UserService] Fetched profile bundle for user %s", userID)
	return response, nil
}

func (s *userServiceImpl) hashPassword(password string) (string, error) {
	// Use bcrypt with cost 12
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// entityToResponse converts User entity to UserResponse
func (s *userServiceImpl) entityToResponse(user *entities.User) *UserResponse {
	resp := &UserResponse{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		RoleID:              user.RoleID,
		IsActive:            user.IsActive,
		IsLocked:            user.IsLocked,
		FailedLoginAttempts: user.FailedLoginAttempts,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
		LastLogin:           user.LastLogin,
	}
	return resp
}
