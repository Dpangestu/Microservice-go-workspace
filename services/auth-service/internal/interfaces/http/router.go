package http

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"bkc_microservice/services/auth-service/internal/application/services"
	shhttp "bkc_microservice/shared/http"
	shmw "bkc_microservice/shared/middleware"
)

func NewRouter(s *services.AuthService) http.Handler {
	r := mux.NewRouter()

	r.Use(shhttp.Recovery)

	rl := shmw.RateLimitTokenPerClient(s.Dep().RDB, 60, time.Minute)

	// r.HandleFunc("/auth/login", LoginHandler(s)).Methods(http.MethodPost)

	r.HandleFunc("/oauth/authorize", MakeAuthorizeHandler(s)).Methods(http.MethodGet, http.MethodPost)
	r.Handle("/oauth/token", rl(http.HandlerFunc(MakeTokenHandler(s)))).Methods(http.MethodPost)
	r.HandleFunc("/oauth/introspect", MakeIntrospectHandler(s)).Methods(http.MethodPost)
	r.HandleFunc("/oauth/revoke", MakeRevokeHandler(s)).Methods(http.MethodPost)

	r.HandleFunc("/oauth/jwks", func(w http.ResponseWriter, _ *http.Request) {
		keys := make([]map[string]string, 0, len(s.Dep().KeyStore.Keys))
		for kid, k := range s.Dep().KeyStore.Keys {
			n := base64.RawURLEncoding.EncodeToString(k.Pub.N.Bytes())
			e := base64.RawURLEncoding.EncodeToString(new(big.Int).SetInt64(int64(k.Pub.E)).Bytes())
			keys = append(keys, map[string]string{
				"kty": "RSA", "use": "sig", "alg": "RS256",
				"n": n, "e": e, "kid": kid,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	}).Methods(http.MethodGet)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	return r
}
