package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"bkc_microservice/services/sync-cbs-service/internal/application/services"
	"bkc_microservice/services/sync-cbs-service/internal/shared"
	shmw "bkc_microservice/shared/middleware"
)

func NewRouter(
	syncService *services.SyncService,
	logger *shared.Logger,
	rdb *redis.Client,
) *mux.Router {
	r := mux.NewRouter()

	syncHandlers := NewSyncCBSHandlers(syncService, logger)

	rl := shmw.RateLimitSlidingWindow(rdb, "rl:sync:api", 100, time.Minute)

	r.HandleFunc("/healthz", syncHandlers.HealthCheck).Methods(http.MethodGet)

	r.Handle(
		"/sync/users/{userID}/input-cbs-data",
		rl(http.HandlerFunc(syncHandlers.InputCBSData)),
	).Methods(http.MethodPost)

	r.Handle(
		"/sync/users/{userID}/mapping",
		rl(http.HandlerFunc(syncHandlers.GetMapping)),
	).Methods(http.MethodGet)

	r.Handle(
		"/sync/mappings/pending",
		rl(http.HandlerFunc(syncHandlers.ListPending)),
	).Methods(http.MethodGet)

	return r
}
