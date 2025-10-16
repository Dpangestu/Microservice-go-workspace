package entities

import "time"

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

type Role struct {
	ID          string
	Name        string
	Description string
	Level       int
	IsSystem    bool
	IsActive    bool
	CreatedAt   time.Time
}

type Permission struct {
	ID          string
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

type User struct {
	ID        string
	CompanyID string
	Username  string
	Email     string
	FirstName string
	LastName  string
	AvatarURL string
	Phone     string
	IsActive  bool
	RoleID    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
