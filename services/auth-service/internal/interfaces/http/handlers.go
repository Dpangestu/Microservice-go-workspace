package http

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"bkc_microservice/services/auth-service/internal/application/services"
)

/* ------------------------------
   /oauth/authorize
------------------------------ */

type authorizeRequest struct {
	ResponseType        string `json:"responseType"`
	ClientID            string `json:"clientId"`
	RedirectURI         string `json:"redirectUri"`
	Scope               string `json:"scope"`
	State               string `json:"state,omitempty"`
	UserID              string `json:"userId"`
	CodeChallenge       string `json:"codeChallenge"`
	CodeChallengeMethod string `json:"codeChallengeMethod"`
	CompanyID           string `json:"companyId"`
}

var consentTmpl = template.Must(template.New("consent").Parse(`
<!doctype html>
<html><head><meta charset="utf-8"><title>Consent</title></head>
<body>
  <h2>App "{{.ClientID}}" minta izin</h2>
  <p>User: {{.UserID}}</p>
  <p>Scope: {{.Scope}}</p>
  <form method="POST" action="/oauth/authorize">
    <input type="hidden" name="response_type" value="code"/>
    <input type="hidden" name="client_id" value="{{.ClientID}}"/>
    <input type="hidden" name="redirect_uri" value="{{.RedirectURI}}"/>
    <input type="hidden" name="scope" value="{{.Scope}}"/>
    <input type="hidden" name="state" value="{{.State}}"/>
    <input type="hidden" name="user_id" value="{{.UserID}}"/>
    <input type="hidden" name="code_challenge" value="{{.CodeChallenge}}"/>
    <input type="hidden" name="code_challenge_method" value="{{.CodeChallengeMethod}}"/>
    <input type="hidden" name="company_id" value="{{.CompanyID}}"/>
    <button type="submit" name="approve" value="1">Allow</button>
    <button type="submit" name="approve" value="0">Deny</button>
  </form>
</body></html>`))

func MakeAuthorizeHandler(s *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req authorizeRequest

		if r.Method == http.MethodGet {
			q := r.URL.Query()
			req = authorizeRequest{
				ResponseType:        q.Get("response_type"),
				ClientID:            q.Get("client_id"),
				RedirectURI:         q.Get("redirect_uri"),
				Scope:               q.Get("scope"),
				State:               q.Get("state"),
				UserID:              q.Get("user_id"),
				CodeChallenge:       q.Get("code_challenge"),
				CodeChallengeMethod: q.Get("code_challenge_method"),
				CompanyID:           q.Get("company_id"),
			}

			if strings.ToLower(q.Get("prompt")) != "none" && q.Get("autoapprove") != "1" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_ = consentTmpl.Execute(w, req)
				return
			}
		} else {
			_ = r.ParseForm()
			req = authorizeRequest{
				ResponseType:        r.FormValue("response_type"),
				ClientID:            r.FormValue("client_id"),
				RedirectURI:         r.FormValue("redirect_uri"),
				Scope:               r.FormValue("scope"),
				State:               r.FormValue("state"),
				UserID:              r.FormValue("user_id"),
				CodeChallenge:       r.FormValue("code_challenge"),
				CodeChallengeMethod: r.FormValue("code_challenge_method"),
				CompanyID:           r.FormValue("company_id"),
			}
			if r.PostFormValue("approve") == "0" {
				http.Error(w, "access_denied", http.StatusForbidden)
				return
			}
		}

		if strings.ToLower(req.ResponseType) != "code" {
			http.Error(w, "unsupported response_type", http.StatusBadRequest)
			return
		}

		code, err := s.StartAuthorizationCode(ctx,
			req.UserID,
			req.ClientID,
			req.RedirectURI,
			req.Scope,
			req.CodeChallenge,
			req.CodeChallengeMethod,
			req.CompanyID)
		if err != nil {
			log.Printf("[/oauth/authorize] userID=%s clientID=%s err=%v", req.UserID, req.ClientID, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, _ := url.Parse(req.RedirectURI)
		q := u.Query()
		q.Set("code", code)
		if req.State != "" {
			q.Set("state", req.State)
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

/* ------------------------------
   /oauth/token
------------------------------ */

type tokenForm struct {
	GrantType    string `json:"grantType"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirectUri,omitempty"`
	CodeVerifier string `json:"codeVerifier,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	Scope        string `json:"scope,omitempty"`
	CompanyID    string `json:"companyId,omitempty"`
}

func MakeTokenHandler(s *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req tokenForm

		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
		} else {
			_ = r.ParseForm()
			req = tokenForm{
				GrantType:    r.FormValue("grant_type"),
				ClientID:     r.FormValue("client_id"),
				ClientSecret: r.FormValue("client_secret"),
				Username:     r.FormValue("username"),
				Password:     r.FormValue("password"),
				Code:         r.FormValue("code"),
				RedirectURI:  r.FormValue("redirect_uri"),
				CodeVerifier: r.FormValue("code_verifier"),
				RefreshToken: r.FormValue("refresh_token"),
				Scope:        r.FormValue("scope"),
				CompanyID:    r.FormValue("company_id"),
			}
		}

		var (
			res *services.TokenResponse
			err error
		)

		switch strings.ToLower(req.GrantType) {
		case "client_credentials":
			res, err = s.IssueClientCredentials(ctx, req.ClientID, req.ClientSecret, req.Scope, req.CompanyID)
		case "password":
			res, err = s.IssuePassword(ctx, req.ClientID, req.ClientSecret, req.Username, req.Password, req.Scope, req.CompanyID)
		case "authorization_code":
			res, err = s.ExchangeAuthorizationCode(ctx, req.ClientID, req.ClientSecret, req.Code, req.RedirectURI, req.CodeVerifier)
		case "refresh_token":
			res, err = s.Refresh(ctx, req.ClientID, req.RefreshToken)
		default:
			http.Error(w, "unsupported grant_type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}
}

/* ------------------------------
   /oauth/introspect
------------------------------ */

type introspectForm struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty"`
}

func MakeIntrospectHandler(s *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = r.ParseForm()

		cid, csec, ok := parseBasicAuth(r)
		if !ok || verifyClient(ctx, s, cid, csec) != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="oauth2"`)
			http.Error(w, "invalid_client", http.StatusUnauthorized)
			return
		}

		req := introspectForm{
			Token:         r.FormValue("token"),
			TokenTypeHint: r.FormValue("token_type_hint"),
		}
		if strings.TrimSpace(req.Token) == "" {
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}

		res, _ := s.Introspect(ctx, req.Token, req.TokenTypeHint, cid)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}
}

/* ------------------------------
   /oauth/revoke
------------------------------ */

type revokeForm struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint,omitempty"`
}

func MakeRevokeHandler(s *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_ = r.ParseForm()

		cid, csec, ok := parseBasicAuth(r)
		if !ok || verifyClient(ctx, s, cid, csec) != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="oauth2"`)
			http.Error(w, "invalid_client", http.StatusUnauthorized)
			return
		}

		req := revokeForm{
			Token:         r.FormValue("token"),
			TokenTypeHint: r.FormValue("token_type_hint"),
		}
		if strings.TrimSpace(req.Token) == "" {
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}

		_ = s.Revoke(ctx, req.Token, req.TokenTypeHint)
		w.WriteHeader(http.StatusOK)
	}
}

/* ------------------------------
   Helpers
------------------------------ */

func parseBasicAuth(r *http.Request) (string, string, bool) {
	ah := r.Header.Get("Authorization")
	if !strings.HasPrefix(ah, "Basic ") {
		return "", "", false
	}
	raw := strings.TrimPrefix(ah, "Basic ")
	dec, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", "", false
	}
	p := strings.SplitN(string(dec), ":", 2)
	if len(p) != 2 {
		return "", "", false
	}
	return p[0], p[1], true
}

func verifyClient(ctx context.Context, s *services.AuthService, clientID, clientSecret string) error {
	c, err := s.Dep().ClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return err
	}
	if c.Secret == nil {
		return errors.New("client has no secret")
	}
	if subtle.ConstantTimeCompare([]byte(*c.Secret), []byte(clientSecret)) != 1 {
		return errors.New("invalid client secret")
	}
	return nil
}
