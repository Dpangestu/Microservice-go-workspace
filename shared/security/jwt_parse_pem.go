package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func ParseTokenWithPEM(token string, pemBytes []byte) (*TokenClaims, *rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, nil, errors.New("bad pem")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	pub := pubIfc.(*rsa.PublicKey)
	claims, err := ParseAndVerify(token, pub)
	return claims, pub, err
}
