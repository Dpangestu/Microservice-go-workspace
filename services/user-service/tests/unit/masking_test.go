package unit

import (
	"testing"

	"bkc_microservice/services/user-service/internal/application/services"

	"github.com/stretchr/testify/assert"
)

func TestMaskUserBundleByScope_BasicScope(t *testing.T) {
	// Arrange
	bundle := map[string]interface{}{
		"data": map[string]interface{}{
			"id":                  "user-123",
			"username":            "testuser",
			"email":               "test@example.com",
			"failedLoginAttempts": 5,
			"isLocked":            true,
		},
		"settings": map[string]string{"theme": "dark"},
	}

	// Act
	result := services.MaskUserBundleByScope(bundle, "profile")

	// Assert
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "user-123", data["id"])
	assert.Nil(t, data["failedLoginAttempts"], "failedLoginAttempts should be masked")
	assert.Nil(t, data["isLocked"], "isLocked should be masked")
	assert.Nil(t, result["settings"], "settings should be masked")
}

func TestMaskUserBundleByScope_AdminScope(t *testing.T) {
	// Arrange
	bundle := map[string]interface{}{
		"data": map[string]interface{}{
			"id":                  "user-123",
			"failedLoginAttempts": 5,
			"isLocked":            true,
		},
	}

	// Act
	result := services.MaskUserBundleByScope(bundle, "profile:sensitive")

	// Assert
	data := result["data"].(map[string]interface{})
	assert.Equal(t, 5, data["failedLoginAttempts"], "failedLoginAttempts should be visible")
	assert.Equal(t, true, data["isLocked"], "isLocked should be visible")
}

func TestMaskUserBundleByScope_MultipleScopes(t *testing.T) {
	// Arrange
	bundle := map[string]interface{}{
		"data":     map[string]interface{}{"id": "user-123"},
		"settings": map[string]string{"theme": "dark"},
	}

	// Act
	result := services.MaskUserBundleByScope(bundle, "profile email openid profile:settings")

	// Assert
	assert.NotNil(t, result["settings"], "settings should be visible with profile:settings scope")
}
