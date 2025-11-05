package fixtures

import (
	"time"

	"bkc_microservice/services/user-service/internal/domain/entities"
)

// UserBuilder untuk membuat test data dengan fluent API
type UserBuilder struct {
	user *entities.User
}

func NewUser() *UserBuilder {
	return &UserBuilder{
		user: &entities.User{
			ID:        "user-test-123",
			Username:  "testuser",
			Email:     "test@example.com",
			RoleID:    1,
			IsActive:  true,
			IsLocked:  false,
			CreatedAt: time.Now(),
		},
	}
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithRoleID(roleID int) *UserBuilder {
	b.user.RoleID = roleID
	return b
}

func (b *UserBuilder) WithIsActive(active bool) *UserBuilder {
	b.user.IsActive = active
	return b
}

func (b *UserBuilder) WithIsLocked(locked bool) *UserBuilder {
	b.user.IsLocked = locked
	return b
}

func (b *UserBuilder) Build() *entities.User {
	return b.user
}

// Profile fixtures
type ProfileBuilder struct {
	profile *entities.UserProfile
}

func NewProfile() *ProfileBuilder {
	return &ProfileBuilder{
		profile: &entities.UserProfile{
			UserID:      "user-test-123",
			FullName:    strPtr("Test User"),
			DisplayName: strPtr("Test"),
			Phone:       strPtr("+6281234567890"),
			Locale:      strPtr("id"),
			Timezone:    strPtr("Asia/Jakarta"),
			CreatedAt:   time.Now(),
		},
	}
}

func (b *ProfileBuilder) WithFullName(name string) *ProfileBuilder {
	b.profile.FullName = &name
	return b
}

func (b *ProfileBuilder) WithTimezone(tz string) *ProfileBuilder {
	b.profile.Timezone = &tz
	return b
}

func (b *ProfileBuilder) Build() *entities.UserProfile {
	return b.profile
}

// Helpers
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
