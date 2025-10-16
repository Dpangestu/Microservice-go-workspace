package http

import (
	"bkc_microservice/services/user-service/internal/application/services"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(s *services.UserService) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/users/{id}", GetUserHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/users", CreateUserHandler(s)).Methods(http.MethodPost)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	return r
}
