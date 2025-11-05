package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/domain/repositories"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	UserRepo         repositories.UserRepository
	RoleRepo         repositories.RoleRepository
	PermissionRepo   repositories.PermissionRepository
	UserActivityRepo repositories.UserActivityRepository
	ProfileRepo      repositories.UserProfileRepository
	SettingsRepo     repositories.UserSettingsRepository
	RPRepo           repositories.RolePermissionsRepository
	RedisClient      *redis.Client // NEW
}

func NewUserService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	userActivityRepo repositories.UserActivityRepository,
	profileRepo repositories.UserProfileRepository,
	settingsRepo repositories.UserSettingsRepository,
	rpRepo repositories.RolePermissionsRepository,
	redisClient *redis.Client, // NEW
) *UserService {
	return &UserService{
		UserRepo:         userRepo,
		RoleRepo:         roleRepo,
		PermissionRepo:   permissionRepo,
		UserActivityRepo: userActivityRepo,
		ProfileRepo:      profileRepo,
		SettingsRepo:     settingsRepo,
		RPRepo:           rpRepo,
		RedisClient:      redisClient, // NEW
	}
}

// Dipakai di handler GetUserHandler
func (s *UserService) GetUserProfile(id string) (*entities.User, error) {
	u, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	// optional: load role & permissions jika perlu
	return u, nil
}

// Tambahan: dipanggil oleh internal handlers yang kita buat
func (s *UserService) FindByEmail(email string) (*entities.User, error) {
	e, err := s.UserRepo.FindByEmail(email)
	log.Println("FindByEmail service:", e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *UserService) FindByID(id string) (*entities.User, error) {
	return s.UserRepo.FindByID(id)
}

func (s *UserService) ListAll() ([]*entities.User, error) {
	// BEFORE:
	// return nil, errors.New("ListAll not implemented")

	// AFTER: Call repository
	return s.UserRepo.ListAll()
}

func (s *UserService) CreateUser(user *entities.User) error {
	return s.UserRepo.Create(user)
}

func (s *UserService) GetUserActivities(userID string) ([]*entities.UserActivity, error) {
	return s.UserActivityRepo.GetByUserID(userID)
}

// GetCurrentUserBundle dengan caching, audit, dan masking
func (s *UserService) GetCurrentUserBundle(ctx context.Context, userID, clientID, scope string, r *http.Request) (map[string]interface{}, error) {
	// 1. TRY CACHE FIRST
	cachedData, err := s.GetCachedUserBundle(ctx, userID)
	if err == nil && cachedData != nil {
		// Apply masking pada cached data
		maskedData := MaskUserBundleByScope(cachedData, scope)

		// Log akses (async, non-blocking)
		s.LogUserProfileAccess(userID, clientID, r)

		return maskedData, nil
	}

	// 2. FETCH FROM DATABASE
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		log.Printf("[UserService] Error fetching user %s: %v", userID, err)
		return nil, errors.New("user not found")
	}

	// 3. BUILD BUNDLE
	dataMap := map[string]interface{}{
		"id":                  user.ID,
		"username":            user.Username,
		"email":               user.Email,
		"isActive":            user.IsActive,
		"isLocked":            user.IsLocked,
		"failedLoginAttempts": user.FailedLoginAttempts,
		"lastLogin":           user.LastLogin,
		"createdAt":           user.CreatedAt,
		"updatedAt":           user.UpdatedAt,
	}
	// Role & permissions
	if user.RoleID > 0 {
		dataMap["roleId"] = user.RoleID
		roleName, _ := s.RPRepo.GetRoleName(user.RoleID)
		if roleName != nil {
			dataMap["roleName"] = *roleName
		}

		// permissions, _ := s.RPRepo.GetPermissionsByRoleID(strconv.Itoa(roleID))
		permissions, _ := s.RPRepo.GetPermissionsByRoleID(user.RoleID)
		if len(permissions) > 0 {
			dataMap["permissions"] = permissions
		}
	}

	// Profile
	profileMap := map[string]interface{}{}
	profile, _ := s.ProfileRepo.FindByUserID(userID)
	if profile != nil {
		if profile.FullName != nil {
			profileMap["fullName"] = *profile.FullName
		}
		if profile.DisplayName != nil {
			profileMap["displayName"] = *profile.DisplayName
		}
		if profile.Phone != nil {
			profileMap["phone"] = *profile.Phone
		}
		if profile.AvatarURL != nil {
			profileMap["avatarUrl"] = *profile.AvatarURL
		}
		if profile.Locale != nil {
			profileMap["locale"] = *profile.Locale
		}
		if profile.Timezone != nil {
			profileMap["timezone"] = *profile.Timezone
		}
		if len(profile.Metadata) > 0 {
			profileMap["metadata"] = profile.Metadata
		}
	}

	// Settings
	settings, _ := s.SettingsRepo.GetByUserID(userID)
	if len(settings) == 0 {
		settings = make(map[string]string)
	}

	// Build response
	resp := map[string]interface{}{
		"data":     dataMap,
		"settings": settings,
		"links": map[string]string{
			"self": "/user/me",
		},
	}

	if len(profileMap) > 0 {
		resp["profile"] = profileMap
	}

	// 4. CACHE BUNDLE (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.CacheUserBundle(ctx, userID, resp)
	}()

	// 5. APPLY MASKING
	resp = MaskUserBundleByScope(resp, scope)

	// 6. LOG AUDIT (async, non-blocking)
	s.LogUserProfileAccess(userID, clientID, r)

	return resp, nil
}
