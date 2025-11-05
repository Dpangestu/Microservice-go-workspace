package http

import (
	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/internal/middleware"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type MeResponse struct {
	Data struct {
		ID                  string     `json:"id"`
		Username            string     `json:"username"`
		Email               string     `json:"email"`
		RoleID              *int       `json:"roleId,omitempty"`
		RoleName            *string    `json:"roleName,omitempty"`
		Permissions         []string   `json:"permissions,omitempty"`
		IsActive            bool       `json:"isActive"`
		IsLocked            bool       `json:"isLocked"`
		FailedLoginAttempts int        `json:"failedLoginAttempts"`
		LastLogin           *time.Time `json:"lastLogin"`
		CreatedAt           time.Time  `json:"createdAt"`
		UpdatedAt           *time.Time `json:"updatedAt"`
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
	Meta     struct {
		TenantID string `json:"tenantId"`
		ClientID string `json:"clientId"`
		Scope    string `json:"scope"`
	} `json:"meta"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func MakeGetCurrentUserHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 1. Extract claims dari middleware
		// claims, ok := middleware.ClaimsFromContext(ctx)
		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok || claims.UserID == "" {
			writeError(w, 401, "unauthorized", "missing or invalid user claims")
			return
		}

		// 2. Ambil bundle dengan caching, audit logging, dan masking
		dataMap, err := s.GetCurrentUserBundle(
			ctx,
			claims.UserID,
			claims.ClientID, // untuk audit
			claims.Scope,    // untuk masking
			r,               // untuk IP & user agent
		)
		if err != nil {
			log.Printf("[Handler /me] Error: %v", err)
			writeError(w, 404, "not_found", "user not found")
			return
		}

		// 3. Add meta
		metaMap := map[string]string{
			"tenantId": claims.TenantID,
			"clientId": claims.ClientID,
			"scope":    claims.Scope,
		}
		dataMap["meta"] = metaMap

		// 4. Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(dataMap); err != nil {
			log.Printf("[Handler /me] Error encoding response: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	errResp := map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errResp)
}

func GetUserHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		user, err := s.GetUserProfile(id)
		fmt.Println("user handler:", user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(user)
	}
}

func CreateUserHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user entities.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.CreateUser(&user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func GetUserActivitiesHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		activities, err := s.GetUserActivities(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(activities)
	}
}
