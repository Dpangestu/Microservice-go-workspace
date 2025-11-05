package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"bkc_microservice/shared/security"
)

type ctxKey string

const claimsKey ctxKey = "jwt_claims"

func ClaimsFromContext(ctx context.Context) (*security.TokenClaims, bool) {
	c, ok := ctx.Value(claimsKey).(*security.TokenClaims)
	return c, ok
}

func RequireJWTWithJWKS(jwks *security.JWKSCache, expectedIssuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			log.Printf("Authorization Header: %s", r.Header.Get("Authorization"))

			if auth == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimSpace(parts[1])

			_, claims, err := jwks.VerifyRS256(token, expectedIssuer)
			if err != nil {
				log.Printf("Error verifying JWT: %v", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			log.Printf("Claims added to context: %+v", claims)

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScopeFromClaims(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || !hasScope(claims.Scope, scope) {
				http.Error(w, "insufficient_scope", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasScope(scopes, want string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == want {
			return true
		}
	}
	return false
}

//
