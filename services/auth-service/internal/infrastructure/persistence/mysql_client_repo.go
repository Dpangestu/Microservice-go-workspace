package persistence

import (
	"context"
	"database/sql"
	"errors"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"
)

type MySQLClientRepo struct{ db *sql.DB }

func NewMySQLClientRepo(db *sql.DB) repositories.ClientRepository {
	return &MySQLClientRepo{db: db}
}

func (r *MySQLClientRepo) FindByClientID(ctx context.Context, clientID string) (*entities.OAuthClient, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, client_id, client_secret, redirect_uri, scopes, company_id, created_at
		FROM oauth_clients WHERE client_id = ?
	`, clientID)

	var c entities.OAuthClient
	var secret, redirect, scopes, companyID sql.NullString

	if err := row.Scan(&c.ID, &c.ClientID, &secret, &redirect, &scopes, &companyID, &c.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Valid {
		c.Secret = &secret.String
	}
	if redirect.Valid {
		c.RedirectURI = &redirect.String
	}
	if scopes.Valid {
		c.Scopes = &scopes.String
	}
	if companyID.Valid {
		c.CompanyID = &companyID.String
	}

	return &c, nil
}
