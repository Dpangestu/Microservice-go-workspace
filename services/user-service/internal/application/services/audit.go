package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// LogUserProfileAccess mencatat setiap akses ke profile
func (s *UserService) LogUserProfileAccess(userID, clientID string, r *http.Request) error {
	// Async non-blocking
	go func() {
		_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		activity := &entities.UserActivity{
			ID:          generateUUID(),
			UserID:      userID,
			Action:      "profile_access",
			Description: fmt.Sprintf("User accessed own profile via %s client", clientID),
			IPAddress:   extractClientIP(r),
			UserAgent:   r.UserAgent(),
			CreatedAt:   time.Now(),
		}

		if err := s.UserActivityRepo.Create(activity); err != nil {
			log.Printf("[AuditLog] Error logging profile access for user %s: %v", userID, err)
		}
	}()

	return nil
}

// extractClientIP mengambil IP dari request (support X-Forwarded-For)
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (saat di belakang proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			// Ambil IP pertama dari daftar X-Forwarded-For dan trim spasi
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback ke RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// generateUUID adalah placeholder - gunakan uuid package
func generateUUID() string {
	// TODO: implement dengan github.com/google/uuid
	return ""
}
