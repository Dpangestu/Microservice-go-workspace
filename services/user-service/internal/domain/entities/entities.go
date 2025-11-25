package entities

import (
	"encoding/json"
	"time"
)

// User represents a user entity
type User struct {
	ID                  string        `json:"id"`
	Username            string        `json:"username"`
	Email               string        `json:"email"`
	PasswordHash        string        `json:"password_hash"`
	RoleID              int           `json:"role_id"`
	IsActive            bool          `json:"is_active"`
	IsLocked            bool          `json:"is_locked"`
	FailedLoginAttempts int           `json:"failed_login_attempts"`
	LastLogin           *time.Time    `json:"last_login"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           *time.Time    `json:"updated_at"`
	Role                *Role         `json:"role"`
	Profile             *UserProfile  `json:"profile"`
	Permissions         []*Permission `json:"permissions"`
}

// Role represents a role entity
type Role struct {
	ID          int
	Name        string
	Description string
	Level       int
	IsActive    bool
	TenantID    *string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// Permission represents a permission entity
type Permission struct {
	ID          int
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// RolePermission represents the relationship between role and permission
type RolePermission struct {
	ID           int
	RoleID       int
	PermissionID int
	CreatedAt    time.Time
}

// UserActivity represents user activity audit logs
type UserActivity struct {
	ID          int
	UserID      string
	Action      string
	Resource    string
	Description string
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

// UserProfile represents additional user profile information
type UserProfile struct {
	ID          int             `json:"id"`
	UserID      string          `json:"userId"`
	FullName    *string         `json:"fullName,omitempty"`
	DisplayName *string         `json:"displayName,omitempty"`
	Phone       *string         `json:"phone,omitempty"`
	AvatarURL   *string         `json:"avatarUrl,omitempty"`
	Locale      *string         `json:"locale,omitempty"`
	Timezone    *string         `json:"timezone,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   *time.Time      `json:"updatedAt"`
}

// UserSettings represents user preference settings
type UserSettings struct {
	ID            int
	UserID        string
	ThemeMode     string
	Language      string
	TwoFAEnabled  bool
	Notifications bool
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// SycCoreUser represents sycrone core user entity
type SycCoreUser struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	UserCore      string    `json:"user_core"`
	KodeGroup1    string    `json:"kode_group_1"`
	KodePerkiraan int       `json:"kode_perkiraan"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Convenience method untuk membuat audit log
func NewUserActivityLog(action, description, ip, ua string) *UserActivity {
	return &UserActivity{
		Action:      action,
		Description: description,
		IPAddress:   ip,
		UserAgent:   ua,
		CreatedAt:   time.Now(),
	}
}
