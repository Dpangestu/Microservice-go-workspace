package security

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kid string `json:"kid"`
}

type jwksPayload struct {
	Keys []jwkKey `json:"keys"`
}

type JWKSCache struct {
	URL       string
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	TTL       time.Duration
	Client    *http.Client
}

func NewJWKSCache(url string, ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		URL:    url,
		TTL:    ttl,
		Client: &http.Client{Timeout: 5 * time.Second},
		keys:   map[string]*rsa.PublicKey{},
	}
}

func (c *JWKSCache) refresh(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad jwks status")
	}
	var p jwksPayload
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}
	m := make(map[string]*rsa.PublicKey)
	for _, k := range p.Keys {
		nb, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eb, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eb {
			e = e<<8 + int(b)
		}
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}
		m[k.Kid] = pub
	}
	c.mu.Lock()
	c.keys = m
	c.expiresAt = time.Now().Add(c.TTL)
	c.mu.Unlock()
	return nil
}

func (c *JWKSCache) keyForKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if time.Now().Before(c.expiresAt) {
		if k, found := c.keys[kid]; found {
			c.mu.RUnlock()
			return k, nil
		}
	}
	c.mu.RUnlock()
	_ = c.refresh(ctx)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if k, found := c.keys[kid]; found {
		return k, nil
	}
	return nil, errors.New("kid not found")
}

func (c *JWKSCache) VerifyRS256(token string, expectedIssuer string) (*jwt.Token, *TokenClaims, error) {
	parser := jwt.Parser{}
	var claims TokenClaims
	tok, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("alg")
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("no kid")
		}
		return c.keyForKid(context.Background(), kid)
	})
	if err != nil {
		return nil, nil, err
	}
	if !tok.Valid {
		return nil, nil, errors.New("invalid")
	}
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return nil, nil, errors.New("iss mismatch")
	}
	return tok, &claims, nil
}
