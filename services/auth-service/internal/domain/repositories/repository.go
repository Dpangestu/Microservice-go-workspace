package repositories

import (
	"context"
	"time"

	"bkc_microservice/services/auth-service/internal/domain/entities"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
	CheckPassword(ctx context.Context, userID string, password string) (bool, error)
}

type ClientRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*entities.OAuthClient, error)
}

type AuthCodeRepository interface {
	Save(ctx context.Context, ac *entities.AuthCode) error
	FindValid(ctx context.Context, code string, now time.Time) (*entities.AuthCode, error)
	DeleteByCode(ctx context.Context, code string) error
}

type TokenRepository interface {
	Save(ctx context.Context, t *entities.Token) error
	FindByRefreshToken(ctx context.Context, refresh string) (*entities.Token, error)
	FindByAccessToken(ctx context.Context, access string) (*entities.Token, error)
	RevokeByAccessToken(ctx context.Context, access string) error
	RevokeByRefreshToken(ctx context.Context, refresh string) error
	CleanupExpired(ctx context.Context, now time.Time) error
}
