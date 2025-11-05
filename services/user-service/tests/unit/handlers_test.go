package unit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"bkc_microservice/services/user-service/internal/application/services"
	httphandlers "bkc_microservice/services/user-service/internal/http" // ← ADD THIS
	"bkc_microservice/services/user-service/tests/fixtures"
	shsec "bkc_microservice/shared/security"

	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	userRepo *fixtures.MockUserRepository
	service  *services.UserService
	handler  http.HandlerFunc
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.userRepo = new(fixtures.MockUserRepository)
	suite.service = services.NewUserService(
		suite.userRepo,
		nil, nil, nil, nil, nil, nil,
		nil,
	)
	suite.handler = httphandlers.MakeGetCurrentUserHandler(suite.service)
}

// Helper function untuk inject claims ke context
func (suite *HandlerTestSuite) createRequestWithClaims(userID, clientID string) *http.Request {
	req := httptest.NewRequest("GET", "/me", nil)
	claims := &shsec.TokenClaims{
		UserID:   userID,
		ClientID: clientID,
		TenantID: "tenant-1",
		Scope:    "profile email",
		Type:     "access",
	}

	// PENTING: Use the EXACT key yang handler expect!
	// Cek di handler code, pakai key yang sama
	ctx := context.WithValue(req.Context(), "jwt_claims", claims)
	req = req.WithContext(ctx)
	return req
}

func (suite *HandlerTestSuite) TestGetUserMe_Success200() {
	// Arrange
	user := fixtures.NewUser().Build()
	suite.userRepo.On("FindByID", user.ID).Return(user, nil)

	req := suite.createRequestWithClaims(user.ID, "mobile")

	// Act
	rr := httptest.NewRecorder()
	suite.handler.ServeHTTP(rr, req)

	// Assert
	suite.Equal(http.StatusOK, rr.Code, "Expected 200 OK, got %d", rr.Code)
	suite.Equal("application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	suite.NoError(err)
	suite.NotNil(response["data"], "Response data should not be nil")
}

func (suite *HandlerTestSuite) TestGetUserMe_Unauthorized() {
	// Arrange
	req := httptest.NewRequest("GET", "/me", nil)
	// NO claims

	// Act
	rr := httptest.NewRecorder()
	suite.handler.ServeHTTP(rr, req)

	// Assert
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *HandlerTestSuite) TestGetUserMe_UserNotFound() {
	// Arrange
	userID := "nonexistent"
	suite.userRepo.On("FindByID", userID).Return(nil, errors.New("not found"))

	req := suite.createRequestWithClaims(userID, "mobile")

	// Act
	rr := httptest.NewRecorder()
	suite.handler.ServeHTTP(rr, req)

	// Assert
	suite.Equal(http.StatusNotFound, rr.Code)
}

func (suite *HandlerTestSuite) TestGetUserMe_ResponseFormat() {
	// Arrange
	user := fixtures.NewUser().Build()
	suite.userRepo.On("FindByID", user.ID).Return(user, nil)

	req := suite.createRequestWithClaims(user.ID, "web")

	// Act
	rr := httptest.NewRecorder()
	suite.handler.ServeHTTP(rr, req)

	// Assert
	suite.Equal(http.StatusOK, rr.Code)

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "data should be map")
	suite.Equal(user.ID, data["id"])

	meta, ok := response["meta"].(map[string]interface{})
	suite.True(ok, "meta should be map")
	suite.Equal("web", meta["clientId"])

	links, ok := response["links"].(map[string]interface{})
	suite.True(ok, "links should be map")
	suite.Equal("/user/me", links["self"])
}

// func TestHandlerTestSuite(t *testing.T) {
// 	suite.Run(t, new(HandlerTestSuite))
// }
