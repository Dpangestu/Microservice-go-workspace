package http

import (
	"encoding/json"
	"log"
	"net/http"

	"bkc_microservice/services/user-service/internal/application/services"

	"github.com/gorilla/mux"
)

/* ==================== INTERNAL ENDPOINTS ==================== */
/* Endpoints ini hanya untuk komunikasi antar-service */

// InternalUserResponse adalah response untuk internal endpoints
type InternalUserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FullName    string `json:"fullName,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Role        string `json:"role,omitempty"`
}

// VerifyInternalAPIKey middleware untuk protect internal endpoints
func VerifyInternalAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key dari header
		apiKey := r.Header.Get("X-Internal-API-Key")

		// Hard-coded untuk demo, sebaiknya dari env variable
		const expectedKey = "shared-secret"

		if apiKey != expectedKey {
			log.Printf("[Internal API] Unauthorized access attempt with key: %s", apiKey)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MakeGetUserByEmailHandler - GET /internal/users/by-email?email=xxx
func MakeGetUserByEmailHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		email := r.URL.Query().Get("email")
		if email == "" {
			http.Error(w, "email parameter required", http.StatusBadRequest)
			return
		}

		log.Printf("[Internal API] GetUserByEmail: %s", email)

		// Call user service to find user
		user, err := s.FindByEmail(email)
		if err != nil {
			log.Printf("[Internal API] Database error 1: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		resp := InternalUserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// MakeGetUserByIDHandler - GET /internal/users/{id}
func MakeGetUserByIDHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := mux.Vars(r)["id"]
		if userID == "" {
			http.Error(w, "user id required", http.StatusBadRequest)
			return
		}

		log.Printf("[Internal API] GetUserByID: %s", userID)

		// Call user service to find user
		user, err := s.FindByID(userID)
		if err != nil {
			log.Printf("[Internal API] Database error 2: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		resp := InternalUserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// MakeListUsersHandler - GET /internal/users (admin only)
func MakeListUsersHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Internal API] ListUsers")

		// Optional: add pagination support
		// page := r.URL.Query().Get("page")
		// limit := r.URL.Query().Get("limit")

		users, err := s.ListAll()
		if err != nil {
			log.Printf("[Internal API] Database error 3: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resp := make([]InternalUserResponse, len(users))
		for i, u := range users {
			resp[i] = InternalUserResponse{
				ID:       u.ID,
				Email:    u.Email,
				Username: u.Username,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

// MakeGetUserProfileHandler - GET /internal/users/{id}/profile
func MakeGetUserProfileHandler(s *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := mux.Vars(r)["id"]
		if userID == "" {
			http.Error(w, "user id required", http.StatusBadRequest)
			return
		}

		log.Printf("[Internal API] GetUserProfile: %s", userID)

		profile, err := s.GetUserProfile(userID)
		if err != nil {
			log.Printf("[Internal API] Database error 4: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if profile == nil {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(profile)
	}
}
