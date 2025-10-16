package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
)

func RequireScopes(publicPEM []byte, required ...string) func(http.Handler) http.Handler {
	req := make(map[string]struct{}, len(required))
	for _, s := range required {
		s = strings.TrimSpace(s)
		if s != "" {
			req[s] = struct{}{}
		}
	}
	pub, err := parseRSAPubFromPEM(publicPEM)
	if err != nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "auth misconfigured", http.StatusInternalServerError)
			})
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer"))
			claims, err := ParseAndVerify(raw, pub)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			if len(req) > 0 {
				h := map[string]struct{}{}
				for _, s := range strings.Fields(claims.Scope) {
					h[s] = struct{}{}
				}
				for s := range req {
					if _, ok := h[s]; !ok {
						w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
						http.Error(w, "insufficient_scope", http.StatusForbidden)
						return
					}
				}
			}
			if claims.UserID != "" {
				r.Header.Set("X-User-Id", claims.UserID)
			}
			if claims.ClientID != "" {
				r.Header.Set("X-Client-Id", claims.ClientID)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseRSAPubFromPEM(b []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("bad pem")
	}
	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not rsa public key")
	}
	return pub, nil
}
