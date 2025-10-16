// D:\Dpangestu\Project\go\bkc_microservice\shared\security\jwt.go
package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RS256Signer struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
	issuer  string
}

func MustNewRS256Signer(privPath, pubPath, issuer string) *RS256Signer {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		panic("invalid private pem")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		panic(err)
	}
	pb, _ := pem.Decode(pubPEM)
	if pb == nil {
		panic("invalid public pem")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(pb.Bytes)
	if err != nil {
		panic(err)
	}
	pub, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		panic("not rsa public key")
	}

	return &RS256Signer{private: priv, public: pub, issuer: issuer}
}

type TokenClaims struct {
	Scope    string   `json:"scope"`
	ClientID string   `json:"clientId,omitempty"`
	UserID   string   `json:"userId,omitempty"`
	Type     string   `json:"typ"` // "access" | "refresh"
	Audience []string `json:"-"`

	// >>> NEW: tenant marker disematkan ke JWT
	TenantID string `json:"tenantId,omitempty"`

	jwt.RegisteredClaims
}

func (s *RS256Signer) Sign(claims TokenClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   subject(claims.UserID, claims.ClientID),
		Audience:  claims.Audience,
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.private)
}

func subject(userID, clientID string) string {
	if userID != "" {
		return "user:" + userID
	}
	return "client:" + clientID
}

func (s *RS256Signer) PublicPEM() ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(s.public)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

func (s *RS256Signer) PublicKey() *rsa.PublicKey {
	return s.public
}

func (s *RS256Signer) RandBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func ParseAndVerify(tokenStr string, pub *rsa.PublicKey) (*TokenClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("method not RS256")
		}
		return pub, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(*TokenClaims); ok && tok.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

type RS256Key struct {
	KID  string
	Priv *rsa.PrivateKey
	Pub  *rsa.PublicKey
}

type RS256KeyStore struct {
	Active string
	Keys   map[string]*RS256Key
	Issuer string
}

func MustLoadKeyStore(activeKid, issuer string, pairs map[string]struct {
	PrivatePath string
	PublicPath  string
}) *RS256KeyStore {
	ks := &RS256KeyStore{Active: activeKid, Issuer: issuer, Keys: map[string]*RS256Key{}}
	for kid, p := range pairs {
		privPEM, err := os.ReadFile(p.PrivatePath)
		if err != nil {
			panic(err)
		}
		block, _ := pem.Decode(privPEM)
		if block == nil {
			panic("invalid private pem")
		}
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}

		pubPEM, err := os.ReadFile(p.PublicPath)
		if err != nil {
			panic(err)
		}
		pb, _ := pem.Decode(pubPEM)
		if pb == nil {
			panic("invalid public pem")
		}
		ifc, err := x509.ParsePKIXPublicKey(pb.Bytes)
		if err != nil {
			panic(err)
		}
		pub, ok := ifc.(*rsa.PublicKey)
		if !ok {
			panic("not rsa public key")
		}

		ks.Keys[kid] = &RS256Key{KID: kid, Priv: priv, Pub: pub}
	}
	if _, ok := ks.Keys[activeKid]; !ok {
		panic("active kid not found")
	}
	return ks
}

func (ks *RS256KeyStore) SignWithActive(claims TokenClaims, ttl time.Duration) (string, error) {
	key := ks.Keys[ks.Active]
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    ks.Issuer,
		Subject:   subject(claims.UserID, claims.ClientID),
		Audience:  claims.Audience,
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = key.KID
	return t.SignedString(key.Priv)
}
