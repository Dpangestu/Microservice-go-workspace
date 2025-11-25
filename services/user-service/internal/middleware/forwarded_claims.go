package middleware

import (
	"context"
	"net/http"
	"strings"

	shsec "bkc_microservice/shared/security"
)

type ctxKey string

const claimsKey ctxKey = "jwt_claims"

func InjectClaimsFromGateway(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &shsec.TokenClaims{
			UserID:   r.Header.Get("X-User-Id"),
			ClientID: r.Header.Get("X-Client-Id"),
			TenantID: r.Header.Get("X-Tenant-Id"),
			Scope:    r.Header.Get("X-Scope"),
			Type:     "access",
		}
		if strings.TrimSpace(c.UserID) == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClaimsFromContext(ctx context.Context) (*shsec.TokenClaims, bool) {
	c, ok := ctx.Value(claimsKey).(*shsec.TokenClaims)
	return c, ok
}

func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFromContext(r.Context())
			if !ok || !hasScope(c.Scope, scope) {
				http.Error(w, "insufficient_scope", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasScope(all, want string) bool {
	for _, s := range strings.Fields(all) {
		if s == want {
			return true
		}
	}
	return false
}
