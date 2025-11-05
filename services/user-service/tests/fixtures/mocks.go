package fixtures

import (
	"bkc_microservice/services/user-service/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

// ===== MOCKS UNTUK REPOSITORIES =====

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(id string) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*entities.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *entities.User) error {
	return m.Called(user).Error(0)
}

func (m *MockUserRepository) Update(user *entities.User) error {
	return m.Called(user).Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	return m.Called(id).Error(0)
}

func (m *MockUserRepository) ListAll() ([]*entities.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

// MockRoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetAll() ([]*entities.Role, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByID(id string) (*entities.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Role), args.Error(1)
}

// MockPermissionRepository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) GetAll() ([]*entities.Permission, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByRoleID(roleID string) ([]*entities.Permission, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Permission), args.Error(1)
}

// MockUserActivityRepository
type MockUserActivityRepository struct {
	mock.Mock
}

func (m *MockUserActivityRepository) Create(activity *entities.UserActivity) error {
	return m.Called(activity).Error(0)
}

func (m *MockUserActivityRepository) GetByUserID(userID string) ([]*entities.UserActivity, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserActivity), args.Error(1)
}

// MockUserProfileRepository
type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) FindByUserID(userID string) (*entities.UserProfile, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) Create(profile *entities.UserProfile) error {
	return m.Called(profile).Error(0)
}

func (m *MockUserProfileRepository) Update(profile *entities.UserProfile) error {
	return m.Called(profile).Error(0)
}

// MockUserSettingsRepository
type MockUserSettingsRepository struct {
	mock.Mock
}

func (m *MockUserSettingsRepository) GetByUserID(userID string) (map[string]string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockUserSettingsRepository) Set(userID, key, value string) error {
	return m.Called(userID, key, value).Error(0)
}

func (m *MockUserSettingsRepository) Delete(userID, key string) error {
	return m.Called(userID, key).Error(0)
}

// MockRolePermissionsRepository - PENTING: Pastikan ini cocok dengan interface!
type MockRolePermissionsRepository struct {
	mock.Mock
}

func (m *MockRolePermissionsRepository) GetPermissionsByRoleID(roleID int) ([]*entities.Permission, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Permission), args.Error(1)
}

func (m *MockRolePermissionsRepository) GetRoleName(roleID int) (*string, error) {
	args := m.Called(roleID)
	return args.Get(0).(*string), args.Error(1)
}
