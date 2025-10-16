package persistence

import (
	"context"
	"database/sql"
	"time"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"
)

type MySQLAuthCodeRepo struct{ db *sql.DB }

func NewMySQLAuthCodeRepo(db *sql.DB) repositories.AuthCodeRepository {
	return &MySQLAuthCodeRepo{db: db}
}

func (r *MySQLAuthCodeRepo) Save(ctx context.Context, ac *entities.AuthCode) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_auth_codes 
			(code, user_id, client_id, code_challenge, code_challenge_method, redirect_uri, scopes, expires_at, company_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ac.Code, ac.UserID, ac.ClientID, ac.CodeChallenge, ac.CodeChallengeMethod, ac.RedirectURI, ac.Scopes, ac.ExpiresAt, ac.CompanyID)
	return err
}

func (r *MySQLAuthCodeRepo) FindValid(ctx context.Context, code string, now time.Time) (*entities.AuthCode, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, user_id, client_id, code_challenge, code_challenge_method, redirect_uri, scopes, expires_at, company_id
		FROM oauth_auth_codes WHERE code = ? AND expires_at > ?
	`, code, now)

	var ac entities.AuthCode
	if err := row.Scan(&ac.ID, &ac.Code, &ac.UserID, &ac.ClientID, &ac.CodeChallenge,
		&ac.CodeChallengeMethod, &ac.RedirectURI, &ac.Scopes, &ac.ExpiresAt, &ac.CompanyID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ac, nil
}

func (r *MySQLAuthCodeRepo) DeleteByCode(ctx context.Context, code string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM oauth_auth_codes WHERE code = ?`, code)
	return err
}
