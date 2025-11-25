package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// LogUserProfileAccess mencatat setiap akses ke /me endpoint
func (s userServiceImpl) LogUserProfileAccess(ctx context.Context, userID, clientID string, r *http.Request) error {
	// Jangan block request jika audit logging gagal
	go func() {
		activity := &entities.UserActivity{
			UserID:      userID,
			Action:      "profile_access",
			Description: fmt.Sprintf("User accessed own profile via %s client", clientID),
			IPAddress:   extractClientIP(r),
			UserAgent:   r.UserAgent(),
			CreatedAt:   time.Now(),
		}

		if err := s.activityRepo.Create(ctx, activity); err != nil {
			log.Printf("[AuditLog] Error logging profile access for user %s: %v", userID, err)
		}
	}()

	return nil
}

// LogProfileUpdate mencatat perubahan profile
func (s userServiceImpl) LogProfileUpdate(ctx context.Context, userID, clientID string, changes map[string]interface{}, r *http.Request) error {
	go func() {
		description := fmt.Sprintf("User updated profile (fields: %v)", getKeys(changes))

		activity := &entities.UserActivity{
			UserID:      userID,
			Action:      "profile_update",
			Description: description,
			IPAddress:   extractClientIP(r),
			UserAgent:   r.UserAgent(),
			CreatedAt:   time.Now(),
		}

		if err := s.activityRepo.Create(ctx, activity); err != nil {
			log.Printf("[AuditLog] Error logging profile update for user %s: %v", userID, err)
		}
	}()

	return nil
}

// Helper functions
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (saat di belakang proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Fallback ke RemoteAddr
	return r.RemoteAddr
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
