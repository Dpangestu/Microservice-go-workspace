package entities

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type OAuthClient struct {
	ID          string
	ClientID    string
	Secret      *string
	RedirectURI *string
	Scopes      *string
	CompanyID   *string
	CreatedAt   time.Time
}

type AuthCode struct {
	ID                  string
	Code                string
	UserID              string
	ClientID            string
	CompanyID           *string
	CodeChallenge       *string
	CodeChallengeMethod *string
	RedirectURI         *string
	Scopes              *string
	ExpiresAt           time.Time
}

type Token struct {
	ID               string
	UserID           *string
	ClientID         string
	AccessToken      string
	RefreshToken     *string
	Scopes           *string
	CompanyID        string
	ExpiresAt        time.Time // access token expiry
	RefreshExpiresAt time.Time // refresh token expiry (baru)
	CreatedAt        time.Time
}
