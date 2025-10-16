package http

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	shsec "bkc_microservice/shared/security"
)

type ctxKey int

const tokenClaimsKey ctxKey = 77

func TokenClaimsFromContext(ctx context.Context) (shsec.TokenClaims, bool) {
	v := ctx.Value(tokenClaimsKey)
	if v == nil {
		return shsec.TokenClaims{}, false
	}
	c, ok := v.(shsec.TokenClaims)
	return c, ok
}

func setTokenClaims(ctx context.Context, c shsec.TokenClaims) context.Context {
	return context.WithValue(ctx, tokenClaimsKey, c)
}

// RequireScopes: verifikasi Bearer JWT (RS256) + semua scope harus ada.
func RequireScopes(pub *rsa.PublicKey, scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "missing_bearer_token", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimSpace(parts[1])
			claims, err := shsec.ParseAndVerify(tokenStr, pub)
			if err != nil {
				http.Error(w, "invalid_token", http.StatusUnauthorized)
				return
			}
			if !hasAllScopes(claims.Scope, scopes) {
				http.Error(w, "insufficient_scope", http.StatusForbidden)
				return
			}
			r = r.WithContext(setTokenClaims(r.Context(), *claims))
			next.ServeHTTP(w, r)
		})
	}
}

func hasAllScopes(scopeStr string, need []string) bool {
	if len(need) == 0 {
		return true
	}
	have := map[string]bool{}
	for _, s := range strings.Fields(scopeStr) {
		have[s] = true
	}
	for _, n := range need {
		if !have[n] {
			return false
		}
	}
	return true
}
