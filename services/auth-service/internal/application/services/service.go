package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"bkc_microservice/services/auth-service/internal/domain/entities"
	"bkc_microservice/services/auth-service/internal/domain/repositories"
	mfa "bkc_microservice/shared/mfa"
	sharedsec "bkc_microservice/shared/security"
	session "bkc_microservice/shared/session"

	"github.com/redis/go-redis/v9"
)

type Dep struct {
	UserRepo   repositories.UserRepository
	ClientRepo repositories.ClientRepository
	CodeRepo   repositories.AuthCodeRepository
	TokenRepo  repositories.TokenRepository
	KeyStore   *sharedsec.RS256KeyStore
	RDB        *redis.Client

	SessionManager *session.Manager
	MFAService     *mfa.Service

	AccessTTL  time.Duration
	RefreshTTL time.Duration
	CodeTTL    time.Duration
}

type AuthService struct{ dep Dep }

func (s *AuthService) Dep() Dep           { return s.dep }
func NewAuthService(dep Dep) *AuthService { return &AuthService{dep: dep} }

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	RefreshToken string `json:"refreshToken,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

/* =================== helpers =================== */

// pilih tenant: explicit > default client
func (s *AuthService) pickCompanyID(explicit string, client *entities.OAuthClient) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if client != nil && client.CompanyID != nil && *client.CompanyID != "" {
		return *client.CompanyID, nil
	}
	return "", errors.New("tenant_required")
}

// string -> *string ("" => nil)
func strptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// *string -> string (nil => "")
func optionalString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

/************** GRANTS **************/

// Client Credentials — tanpa refresh token
func (s *AuthService) IssueClientCredentials(ctx context.Context, clientID, clientSecret, scope, companyID string) (*TokenResponse, error) {
	c, err := s.dep.ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if c.Secret == nil {
		return nil, errors.New("client has no secret")
	}
	if subtle.ConstantTimeCompare([]byte(*c.Secret), []byte(clientSecret)) != 1 {
		return nil, errors.New("invalid client secret")
	}

	compID, err := s.pickCompanyID(companyID, c)
	if err != nil {
		return nil, err
	}

	at, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		Type:     "access",
		Audience: []string{c.ClientID},
		TenantID: compID,
	}, s.dep.AccessTTL)
	if err != nil {
		return nil, err
	}

	if err := s.dep.TokenRepo.Save(ctx, &entities.Token{
		UserID:           nil,
		ClientID:         c.ID,
		AccessToken:      at,
		RefreshToken:     nil,
		Scopes:           &scope,
		ExpiresAt:        time.Now().Add(s.dep.AccessTTL),
		RefreshExpiresAt: time.Time{},
		CompanyID:        compID,
	}); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken: at,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.dep.AccessTTL.Seconds()),
		Scope:       scope,
	}, nil
}

// Resource Owner Password Credentials (dev/internal)
func (s *AuthService) IssuePassword(ctx context.Context, clientID, clientSecret, username, password, scope, companyID string) (*TokenResponse, error) {
	c, err := s.dep.ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if c == nil {
		return nil, errors.New("invalid client 1")
	}

	if c.Secret != nil && subtle.ConstantTimeCompare([]byte(*c.Secret), []byte(clientSecret)) != 1 {
		return nil, errors.New("invalid client 2")
	}

	u, err := s.dep.UserRepo.FindByEmail(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials 1")
	}
	ok, err := s.dep.UserRepo.CheckPassword(ctx, u.ID, password)
	if err != nil || !ok {
		return nil, errors.New("invalid credentials 2")
	}

	compID, err := s.pickCompanyID(companyID, c)
	if err != nil {
		return nil, err
	}

	at, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		UserID:   u.ID,
		Type:     "access",
		Audience: []string{c.ClientID},
		TenantID: compID,
	}, s.dep.AccessTTL)
	if err != nil {
		return nil, err
	}

	rt, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		UserID:   u.ID,
		Type:     "refresh",
		Audience: []string{c.ClientID},
		TenantID: compID,
	}, s.dep.RefreshTTL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.dep.TokenRepo.Save(ctx, &entities.Token{
		UserID:           &u.ID,
		ClientID:         c.ID,
		AccessToken:      at,
		RefreshToken:     &rt,
		Scopes:           &scope,
		ExpiresAt:        now.Add(s.dep.AccessTTL),
		RefreshExpiresAt: now.Add(s.dep.RefreshTTL),
		CompanyID:        compID,
	}); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  at,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.dep.AccessTTL.Seconds()),
		RefreshToken: rt,
		Scope:        scope,
	}, nil
}

func (s *AuthService) StartAuthorizationCode(ctx context.Context, userID, clientID, redirectURI, scope, codeChallenge, codeMethod, companyID string) (string, error) {
	c, err := s.dep.ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return "", err
	}

	if c.RedirectURI != nil {
		if redirectURI == "" {
			redirectURI = *c.RedirectURI
		} else if redirectURI != *c.RedirectURI {
			return "", errors.New("invalid redirect_uri")
		}
	} else if redirectURI == "" {
		return "", errors.New("redirect_uri required")
	}

	compID, err := s.pickCompanyID(companyID, c)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	code := fmt.Sprintf("%x", buf)

	var cc, cm, ru, sc *string
	if codeChallenge != "" {
		cc = &codeChallenge
	}
	if codeMethod != "" {
		cm = &codeMethod
	}
	if redirectURI != "" {
		ru = &redirectURI
	}
	if scope != "" {
		sc = &scope
	}

	ac := &entities.AuthCode{
		Code:                code,
		UserID:              userID,
		ClientID:            c.ID,
		CodeChallenge:       cc,
		CodeChallengeMethod: cm,
		RedirectURI:         ru,
		Scopes:              sc,
		ExpiresAt:           time.Now().Add(s.dep.CodeTTL),
		CompanyID:           strptr(compID),
	}
	if err := s.dep.CodeRepo.Save(ctx, ac); err != nil {
		return "", err
	}

	log.Printf(
		"[AuthService] Saved auth code: code=%s user_id=%s client_id=%s tenant=%s expires_at=%s",
		ac.Code, ac.UserID, ac.ClientID, optionalString(ac.CompanyID), ac.ExpiresAt.Format(time.RFC3339),
	)
	return code, nil
}

func (s *AuthService) ExchangeAuthorizationCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	c, err := s.dep.ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if c.Secret != nil && subtle.ConstantTimeCompare([]byte(*c.Secret), []byte(clientSecret)) != 1 {
		return nil, errors.New("invalid client")
	}

	ac, err := s.dep.CodeRepo.FindValid(ctx, code, time.Now())
	if err != nil {
		return nil, errors.New("invalid code")
	}

	if ac.RedirectURI != nil && redirectURI != *ac.RedirectURI {
		return nil, errors.New("redirect_uri mismatch")
	}

	if ac.CodeChallenge != nil {
		method := "PLAIN"
		if ac.CodeChallengeMethod != nil {
			method = strings.ToUpper(*ac.CodeChallengeMethod)
		}
		computed := codeVerifier
		if method == "S256" {
			computed = pkceS256(codeVerifier)
		}
		if subtle.ConstantTimeCompare([]byte(computed), []byte(*ac.CodeChallenge)) != 1 {
			return nil, errors.New("invalid code_verifier")
		}
	}

	scope := ""
	if ac.Scopes != nil {
		scope = *ac.Scopes
	}

	tenant := optionalString(ac.CompanyID)

	at, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		UserID:   ac.UserID,
		Type:     "access",
		Audience: []string{c.ClientID},
		TenantID: tenant,
	}, s.dep.AccessTTL)
	if err != nil {
		return nil, err
	}

	rt, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		UserID:   ac.UserID,
		Type:     "refresh",
		Audience: []string{c.ClientID},
		TenantID: tenant,
	}, s.dep.RefreshTTL)
	if err != nil {
		return nil, err
	}

	_ = s.dep.CodeRepo.DeleteByCode(ctx, code)

	now := time.Now()
	if err := s.dep.TokenRepo.Save(ctx, &entities.Token{
		UserID:           &ac.UserID,
		ClientID:         c.ID,
		AccessToken:      at,
		RefreshToken:     &rt,
		Scopes:           &scope,
		ExpiresAt:        now.Add(s.dep.AccessTTL),
		RefreshExpiresAt: now.Add(s.dep.RefreshTTL),
		CompanyID:        tenant,
	}); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  at,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.dep.AccessTTL.Seconds()),
		RefreshToken: rt,
		Scope:        scope,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, clientID, refreshToken string) (*TokenResponse, error) {
	tok, err := s.dep.TokenRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh_token")
	}
	if tok == nil {
		return nil, errors.New("refresh_token_not_found")
	}

	c, err := s.dep.ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, errors.New("invalid client_id")
	}
	if tok.ClientID != c.ID {
		return nil, errors.New("invalid client_id")
	}

	// cek refresh expiry
	refreshDeadline := tok.RefreshExpiresAt
	if refreshDeadline.IsZero() {
		refreshDeadline = tok.CreatedAt.Add(s.dep.RefreshTTL)
	}
	if time.Now().After(refreshDeadline) {
		return nil, errors.New("refresh_token_expired")
	}

	scope := ""
	if tok.Scopes != nil {
		scope = *tok.Scopes
	}

	tenant := tok.CompanyID

	at, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		Type:     "access",
		Audience: []string{c.ClientID},
		TenantID: tenant,
	}, s.dep.AccessTTL)
	if err != nil {
		return nil, err
	}

	newRT, err := s.dep.KeyStore.SignWithActive(sharedsec.TokenClaims{
		Scope:    scope,
		ClientID: c.ClientID,
		Type:     "refresh",
		Audience: []string{c.ClientID},
		TenantID: tenant,
	}, s.dep.RefreshTTL)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.dep.TokenRepo.Save(ctx, &entities.Token{
		UserID:           tok.UserID,
		ClientID:         c.ID,
		AccessToken:      at,
		RefreshToken:     &newRT,
		Scopes:           &scope,
		ExpiresAt:        now.Add(s.dep.AccessTTL),
		RefreshExpiresAt: now.Add(s.dep.RefreshTTL),
		CompanyID:        tenant,
	}); err != nil {
		return nil, err
	}

	_ = s.dep.TokenRepo.RevokeByRefreshToken(ctx, refreshToken)

	return &TokenResponse{
		AccessToken:  at,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.dep.AccessTTL.Seconds()),
		RefreshToken: newRT,
		Scope:        scope,
	}, nil
}

/************** INTROSPECT / REVOKE **************/

type IntrospectionResult struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"clientId,omitempty"`
	Username  string   `json:"username,omitempty"`
	TokenType string   `json:"tokenType,omitempty"`
	Exp       *int64   `json:"exp,omitempty"`
	Iat       *int64   `json:"iat,omitempty"`
	Sub       string   `json:"sub,omitempty"`
	Aud       []string `json:"aud,omitempty"`
}

func (s *AuthService) Introspect(ctx context.Context, token, tokenTypeHint, callerClientPublicID string) (*IntrospectionResult, error) {
	if s.isBlacklisted(ctx, token) {
		return &IntrospectionResult{Active: false}, nil
	}

	if tokenTypeHint == "" || strings.EqualFold(tokenTypeHint, "access_token") {
		if t, err := s.dep.TokenRepo.FindByAccessToken(ctx, token); err == nil && t != nil {
			if callerClientPublicID != "" {
				c, err := s.dep.ClientRepo.FindByClientID(ctx, callerClientPublicID)
				if err != nil || c == nil || t.ClientID != c.ID {
					return &IntrospectionResult{Active: false}, nil
				}
			}

			now := time.Now()
			active := now.Before(t.ExpiresAt)
			scope := optionalString(t.Scopes)

			var exp, iat *int64
			if !t.ExpiresAt.IsZero() {
				e := t.ExpiresAt.Unix()
				exp = &e
			}
			if !t.CreatedAt.IsZero() {
				i := t.CreatedAt.Unix()
				iat = &i
			}
			sub := "client:" + t.ClientID
			if t.UserID != nil && *t.UserID != "" {
				sub = "user:" + *t.UserID
			}

			return &IntrospectionResult{
				Active:    active,
				Scope:     scope,
				ClientID:  t.ClientID,
				Username:  sub,
				TokenType: "access",
				Exp:       exp,
				Iat:       iat,
				Sub:       sub,
				Aud:       []string{t.ClientID},
			}, nil
		}
	}

	if tokenTypeHint == "" || strings.EqualFold(tokenTypeHint, "refresh_token") {
		if t, err := s.dep.TokenRepo.FindByRefreshToken(ctx, token); err == nil && t != nil {
			if callerClientPublicID != "" {
				c, err := s.dep.ClientRepo.FindByClientID(ctx, callerClientPublicID)
				if err != nil || c == nil || t.ClientID != c.ID {
					return &IntrospectionResult{Active: false}, nil
				}
			}

			deadline := t.RefreshExpiresAt
			if deadline.IsZero() {
				deadline = t.CreatedAt.Add(s.dep.RefreshTTL)
			}
			active := time.Now().Before(deadline)
			scope := optionalString(t.Scopes)

			var exp, iat *int64
			if !t.CreatedAt.IsZero() {
				i := t.CreatedAt.Unix()
				iat = &i
			}
			if !deadline.IsZero() {
				e := deadline.Unix()
				exp = &e
			}
			sub := "client:" + t.ClientID
			if t.UserID != nil && *t.UserID != "" {
				sub = "user:" + *t.UserID
			}

			return &IntrospectionResult{
				Active:    active,
				Scope:     scope,
				ClientID:  t.ClientID,
				Username:  sub,
				TokenType: "refresh",
				Exp:       exp,
				Iat:       iat,
				Sub:       sub,
				Aud:       []string{t.ClientID},
			}, nil
		}
	}

	return &IntrospectionResult{Active: false}, nil
}

func (s *AuthService) Revoke(ctx context.Context, token, tokenTypeHint string) error {
	if strings.EqualFold(tokenTypeHint, "refresh_token") {
		_ = s.dep.TokenRepo.RevokeByRefreshToken(ctx, token)
		return nil
	}

	if err := s.dep.TokenRepo.RevokeByAccessToken(ctx, token); err == nil {
		return nil
	}

	_ = s.dep.TokenRepo.RevokeByRefreshToken(ctx, token)

	if strings.EqualFold(tokenTypeHint, "access_token") {
		if t, err := s.dep.TokenRepo.FindByAccessToken(ctx, token); err == nil && t != nil {
			ttl := time.Until(t.ExpiresAt)
			if ttl < 0 {
				ttl = 0
			}
			s.blacklistAccess(ctx, token, ttl)
		}
	}
	return nil
}

func pkceS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (s *AuthService) blacklistAccess(ctx context.Context, token string, ttl time.Duration) {
	if s.dep.RDB == nil {
		return
	}
	_ = s.dep.RDB.Set(ctx, "bl:at:"+token, "1", ttl).Err()
}

func (s *AuthService) isBlacklisted(ctx context.Context, token string) bool {
	if s.dep.RDB == nil {
		return false
	}
	v, _ := s.dep.RDB.Exists(ctx, "bl:at:"+token).Result()
	return v == 1
}
