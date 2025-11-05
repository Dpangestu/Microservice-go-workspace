package persistence

import (
	"context"
	"database/sql"
	"time"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"
)

type MySQLTokenRepo struct{ db *sql.DB }

func NewMySQLTokenRepo(db *sql.DB) repositories.TokenRepository {
	return &MySQLTokenRepo{db: db}
}

func (r *MySQLTokenRepo) Save(ctx context.Context, t *entities.Token) error {
	// 1) insert access token
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_access_tokens
		  (id, user_id, client_id, token, scopes, company_id, expires_at, created_at, revoked)
		VALUES
		  (UUID(), ?, ?, ?, ?, ?, ?, NOW(), 0)
	`, t.UserID, t.ClientID, t.AccessToken, t.Scopes, t.CompanyID, t.ExpiresAt)
	if err != nil {
		return err
	}

	// 2) ambil id access token via hash
	var atID string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id FROM oauth_access_tokens
		WHERE token_sha = UNHEX(SHA2(?,256))
		LIMIT 1
	`, t.AccessToken).Scan(&atID); err != nil {
		return err
	}

	// 3) insert refresh token (jika ada)
	if t.RefreshToken != nil && *t.RefreshToken != "" {
		var rexp *time.Time
		if !t.RefreshExpiresAt.IsZero() {
			rexp = &t.RefreshExpiresAt
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO oauth_refresh_tokens
			  (id, access_token_id, token, company_id, expires_at, created_at, revoked)
			VALUES
			  (UUID(), ?, ?, ?, ?, NOW(), 0)
		`, atID, *t.RefreshToken, t.CompanyID, rexp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MySQLTokenRepo) FindByRefreshToken(ctx context.Context, refresh string) (*entities.Token, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT at.id, at.user_id, at.client_id,
		       at.token      AS access_token,
		       rt.token      AS refresh_token,
		       at.scopes,
		       at.expires_at,
		       rt.expires_at AS refresh_expires_at,
		       at.company_id,
		       at.created_at
		FROM oauth_refresh_tokens rt
		JOIN oauth_access_tokens  at ON at.id = rt.access_token_id
		WHERE rt.token_sha = UNHEX(SHA2(?,256))
		  AND rt.revoked = 0
		  AND at.revoked = 0
	`, refresh)

	var t entities.Token
	var refreshExp sql.NullTime
	if err := row.Scan(&t.ID, &t.UserID, &t.ClientID,
		&t.AccessToken, &t.RefreshToken, &t.Scopes, &t.ExpiresAt, &refreshExp,
		&t.CompanyID, &t.CreatedAt,
	); err != nil {
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

func (r *MySQLTokenRepo) FindByAccessToken(ctx context.Context, access string) (*entities.Token, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT at.id, at.user_id, at.client_id,
		       at.token    AS access_token,
		       NULL        AS refresh_token,
		       at.scopes,
		       at.expires_at,
		       NULL        AS refresh_expires_at,
		       at.company_id,
		       at.created_at
		FROM oauth_access_tokens at
		WHERE at.token_sha = UNHEX(SHA2(?,256))
		  AND at.revoked = 0
	`, access)

	var t entities.Token
	var refreshExp sql.NullTime
	if err := row.Scan(&t.ID, &t.UserID, &t.ClientID,
		&t.AccessToken, &t.RefreshToken, &t.Scopes, &t.ExpiresAt, &refreshExp,
		&t.CompanyID, &t.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *MySQLTokenRepo) RevokeByAccessToken(ctx context.Context, access string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE oauth_access_tokens
		SET revoked = 1
		WHERE token_sha = UNHEX(SHA2(?,256))
	`, access)
	return err
}

func (r *MySQLTokenRepo) RevokeByRefreshToken(ctx context.Context, refresh string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE oauth_refresh_tokens
		SET revoked = 1
		WHERE token_sha = UNHEX(SHA2(?,256))
	`, refresh)
	return err
}

func (r *MySQLTokenRepo) CleanupExpired(ctx context.Context, now time.Time) error {
	// hapus refresh token kedaluwarsa
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM oauth_refresh_tokens
		WHERE expires_at IS NOT NULL AND expires_at < ?
	`, now); err != nil {
		return err
	}
	// hapus access token kedaluwarsa (akan CASCADE refresh via FK)
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM oauth_access_tokens
		WHERE expires_at < ?
	`, now)
	return err
}
