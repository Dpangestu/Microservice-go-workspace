// internal/domain/services/password_service.go
package services

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

type passwordServiceImpl struct{}

func NewPasswordService() PasswordService {
	return &passwordServiceImpl{}
}

// HashPassword returns bcrypt hash of password
func (ps *passwordServiceImpl) HashPassword(password string) (string, error) {
	// bcrypt cost: 10 = good balance between security & speed
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares password with hash
func (ps *passwordServiceImpl) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
