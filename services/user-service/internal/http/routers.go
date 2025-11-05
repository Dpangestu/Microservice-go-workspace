package http

import (
	"net/http"
	"time"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/middleware"
	shmiddleware "bkc_microservice/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func NewRouter(s *services.UserService, rdb *redis.Client) http.Handler {
	r := mux.NewRouter()

	// Health check (no auth)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	// Authenticated routes
	authenticatedRouter := r.PathPrefix("/").Subrouter()
	authenticatedRouter.Use(middleware.InjectClaimsFromGateway)

	// GET /me dengan rate limiting (60 req/min per user)
	authenticatedRouter.Handle("/me",
		shmiddleware.RateLimitUserMeEndpoint(rdb, 60, 1*time.Minute)(
			http.HandlerFunc(MakeGetCurrentUserHandler(s)),
		),
	).Methods(http.MethodGet)

	// Other routes
	authenticatedRouter.HandleFunc("/users/{id}", GetUserHandler(s)).Methods(http.MethodGet)
	authenticatedRouter.HandleFunc("/users", CreateUserHandler(s)).Methods(http.MethodPost)
	authenticatedRouter.HandleFunc("/users/{id}/activities", GetUserActivitiesHandler(s)).Methods(http.MethodGet)

	return r
}
