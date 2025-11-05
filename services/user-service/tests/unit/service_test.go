package unit

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"bkc_microservice/services/user-service/internal/application/services"
	"bkc_microservice/services/user-service/internal/domain/entities"
	"bkc_microservice/services/user-service/tests/fixtures"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	userRepo         *fixtures.MockUserRepository
	roleRepo         *fixtures.MockRoleRepository
	permissionRepo   *fixtures.MockPermissionRepository
	userActivityRepo *fixtures.MockUserActivityRepository
	profileRepo      *fixtures.MockUserProfileRepository
	settingsRepo     *fixtures.MockUserSettingsRepository
	rpRepo           *fixtures.MockRolePermissionsRepository // ← Mock, bukan concrete
	service          *services.UserService
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.userRepo = new(fixtures.MockUserRepository)
	suite.roleRepo = new(fixtures.MockRoleRepository)
	suite.permissionRepo = new(fixtures.MockPermissionRepository)
	suite.userActivityRepo = new(fixtures.MockUserActivityRepository)
	suite.profileRepo = new(fixtures.MockUserProfileRepository)
	suite.settingsRepo = new(fixtures.MockUserSettingsRepository)
	suite.rpRepo = new(fixtures.MockRolePermissionsRepository)

	// Create service - NOW accept interface, not concrete type
	suite.service = services.NewUserService(
		suite.userRepo,
		suite.roleRepo,
		suite.permissionRepo,
		suite.userActivityRepo,
		suite.profileRepo,
		suite.settingsRepo,
		suite.rpRepo, // ← Mock implement interface, so it works!
		nil,          // no Redis
	)
}

func (suite *ServiceTestSuite) TestGetCurrentUserBundle_Success() {
	// Arrange
	user := fixtures.NewUser().Build()
	profile := fixtures.NewProfile().Build()
	roleName := "admin"

	// CHANGE FROM:
	// permissions := []string{"users:read", "users:write"}

	// CHANGE TO:
	permissions := []*entities.Permission{
		{ID: "p1", Name: "users:read"},
		{ID: "p2", Name: "users:write"},
	}

	settings := map[string]string{"theme": "dark"}

	// Mock setup
	suite.userRepo.On("FindByID", user.ID).Return(user, nil)
	suite.rpRepo.On("GetRoleName", 1).Return(&roleName, nil)
	suite.rpRepo.On("GetPermissionsByRoleID", 1).Return(permissions, nil) // ✅ Now []*Permission
	suite.profileRepo.On("FindByUserID", user.ID).Return(profile, nil)
	suite.settingsRepo.On("GetByUserID", user.ID).Return(settings, nil)
	suite.userActivityRepo.On("Create", mock.Anything).Return(nil)

	// Act
	ctx := context.Background()
	req := httptest.NewRequest("GET", "/me", nil)
	result, err := suite.service.GetCurrentUserBundle(ctx, user.ID, "web", "profile", req)

	// Assert
	suite.NoError(err, "Error should be nil")
	suite.NotNil(result, "Result should not be nil")

	data := result["data"].(map[string]interface{})
	suite.Equal(user.ID, data["id"])
	suite.Equal("testuser", data["username"])
}

func (suite *ServiceTestSuite) TestGetCurrentUserBundle_UserNotFound() {
	suite.userRepo.On("FindByID", "nonexistent").Return(nil, errors.New("user not found"))

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/me", nil)
	result, err := suite.service.GetCurrentUserBundle(ctx, "nonexistent", "web", "profile", req)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *ServiceTestSuite) TestFindByID_Success() {
	user := fixtures.NewUser().WithID("user-abc").Build()
	suite.userRepo.On("FindByID", "user-abc").Return(user, nil)

	result, err := suite.service.FindByID("user-abc")

	suite.NoError(err)
	suite.Equal("user-abc", result.ID)
}

func (suite *ServiceTestSuite) TestFindByEmail_Success() {
	user := fixtures.NewUser().Build()
	suite.userRepo.On("FindByEmail", "test@example.com").Return(user, nil)

	result, err := suite.service.FindByEmail("test@example.com")

	suite.NoError(err)
	suite.Equal("test@example.com", result.Email)
}

func (suite *ServiceTestSuite) TestListAll_Success() {
	users := []*entities.User{
		fixtures.NewUser().WithID("1").Build(),
		fixtures.NewUser().WithID("2").Build(),
	}
	suite.userRepo.On("ListAll").Return(users, nil)

	result, err := suite.service.ListAll()

	suite.NoError(err)
	suite.Equal(2, len(result))
}

func (suite *ServiceTestSuite) TestCreateUser_Success() {
	user := fixtures.NewUser().Build()
	suite.userRepo.On("Create", user).Return(nil)

	err := suite.service.CreateUser(user)

	suite.NoError(err)
	suite.userRepo.AssertCalled(suite.T(), "Create", user)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
