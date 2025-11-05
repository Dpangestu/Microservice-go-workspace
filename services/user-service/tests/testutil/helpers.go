package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	shsec "bkc_microservice/shared/security"
)

// CreateTestRequestWithClaims membuat request dengan claims pre-filled
func CreateTestRequestWithClaims(userID, clientID, tenantID, scope string) *http.Request {
	req := httptest.NewRequest("GET", "/me", nil)

	claims := &shsec.TokenClaims{
		UserID:   userID,
		ClientID: clientID,
		TenantID: tenantID,
		Scope:    scope,
		Type:     "access",
	}

	ctx := context.WithValue(req.Context(), "jwt_claims", claims)
	return req.WithContext(ctx)
}

// AssertOKStatus helper untuk assert 200
func AssertOKStatus(t *testing.T, statusCode int) bool {
	t.Helper()
	if statusCode == 200 {
		return true
	}
	t.Errorf("Expected status 200, got %d", statusCode)
	return false
}
