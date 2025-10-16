package persistence

import (
	"context"
	"database/sql"
	"log"
	"time"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"
)

type MySQLTokenRepo struct{ db *sql.DB }

func NewMySQLTokenRepo(db *sql.DB) repositories.TokenRepository {
	return &MySQLTokenRepo{db: db}
}

func (r *MySQLTokenRepo) Save(ctx context.Context, t *entities.Token) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_tokens
			(user_id, client_id, access_token, refresh_token, scopes, expires_at, refresh_expires_at, company_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, t.UserID, t.ClientID, t.AccessToken, t.RefreshToken, t.Scopes, t.ExpiresAt, nullTime(t.RefreshExpiresAt), t.CompanyID)
	return err
}

func (r *MySQLTokenRepo) FindByRefreshToken(ctx context.Context, refresh string) (*entities.Token, error) {
	log.Printf("[FindByRefreshToken] searching for refresh_token len=%d prefix=%s...",
		len(refresh), refresh[:min(30, len(refresh))])

	row := r.db.QueryRowContext(ctx, `
        SELECT id, user_id, client_id, access_token, refresh_token, scopes, expires_at, refresh_expires_at, company_id, created_at
        FROM oauth_tokens WHERE refresh_token = ?
    `, refresh)

	var t entities.Token
	var refreshExp sql.NullTime
	if err := row.Scan(&t.ID, &t.UserID, &t.ClientID, &t.AccessToken, &t.RefreshToken,
		&t.Scopes, &t.ExpiresAt, &refreshExp, &t.CompanyID, &t.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var refreshLen int
	var refreshPrefix string
	if t.RefreshToken != nil {
		refreshLen = len(*t.RefreshToken)
		refreshPrefix = (*t.RefreshToken)[:min(30, len(*t.RefreshToken))]
	}
	log.Printf("[FindByRefreshToken] found token id=%s len=%d prefix=%s",
		t.ID, refreshLen, refreshPrefix)

	if refreshExp.Valid {
		t.RefreshExpiresAt = refreshExp.Time
	}
	return &t, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *MySQLTokenRepo) FindByAccessToken(ctx context.Context, access string) (*entities.Token, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, client_id, access_token, refresh_token, scopes, expires_at, refresh_expires_at, company_id, created_at
		FROM oauth_tokens WHERE access_token = ?
	`, access)

	var t entities.Token
	var refreshExp sql.NullTime
	if err := row.Scan(&t.ID, &t.UserID, &t.ClientID, &t.AccessToken, &t.RefreshToken,
		&t.Scopes, &t.ExpiresAt, &refreshExp, &t.CompanyID, &t.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if refreshExp.Valid {
		t.RefreshExpiresAt = refreshExp.Time
	}
	return &t, nil
}

func (r *MySQLTokenRepo) RevokeByAccessToken(ctx context.Context, access string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE access_token = ?`, access)
	return err
}

func (r *MySQLTokenRepo) RevokeByRefreshToken(ctx context.Context, refresh string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE refresh_token = ?`, refresh)
	return err
}

func (r *MySQLTokenRepo) CleanupExpired(ctx context.Context, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_tokens WHERE expires_at < ?`, now)
	return err
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
