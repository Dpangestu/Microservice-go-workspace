package mfa

import (
	"time"

	"github.com/pquerna/otp/totp"
)

type TOTPService struct{}

// GenerateSecret untuk membuat secret dan otpauth URL (buat di-scan pakai Google Authenticator)
func (t *TOTPService) GenerateSecret(account, issuer string) (secret, url string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// Validate mengecek MFA code dari user
func (t *TOTPService) Validate(passcode, secret string) bool {
	return totp.Validate(passcode, secret)
}

// GenerateCode untuk testing / auto flow (misal admin)
func (t *TOTPService) GenerateCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
