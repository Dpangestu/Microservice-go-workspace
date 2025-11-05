package entities

import (
	"encoding/json"
	"time"
)

type Company struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Address   string
	Website   string
	LogoURL   string
	Timezone  string
	Currency  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type Tenant struct {
	ID        string
	Name      string
	Status    string // Misalnya 'active', 'inactive'
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type Role struct {
	ID          string
	TenantID    string // Tenant terkait dengan role ini
	Name        string
	Description string
	Level       int  // Menentukan tingkat akses (misalnya admin, user)
	IsSystem    bool // Apakah role ini adalah role sistem
	IsActive    bool // Status role aktif atau tidak
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type Permission struct {
	ID          string
	Name        string
	Resource    string // Misalnya: "users", "products", dll
	Action      string // Aksi pada resource, misalnya "create", "read", "update", "delete"
	Description string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type User struct {
	ID                  string        `json:"id"`
	Username            string        `json:"username"`
	Email               string        `json:"email"`
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

type UserProfile struct {
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

type UserActivity struct {
	ID          string
	UserID      string // Menghubungkan aktivitas dengan user
	Action      string // Misalnya: "login", "create_user", dll
	Description string
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}
