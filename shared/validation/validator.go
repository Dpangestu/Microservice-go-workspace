package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator handles request validation
type Validator struct{}

// NewValidator creates new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail checks email format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// ValidatePassword checks password strength
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Check complexity: min uppercase, lowercase, digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("password must contain uppercase, lowercase, and digit")
	}

	return nil
}

// ValidateUsername checks username format
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}

	if len(username) > 50 {
		return fmt.Errorf("username must not exceed 50 characters")
	}

	pattern := `^[a-zA-Z0-9_-]+$`
	matched, _ := regexp.MatchString(pattern, username)
	if !matched {
		return fmt.Errorf("username can only contain letters, numbers, underscore, dash")
	}

	return nil
}

// ValidationError wrapper untuk multiple errors
type ValidationErrors struct {
	Errors map[string]string
}

func (ve *ValidationErrors) Error() string {
	var msgs []string
	for field, msg := range ve.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(msgs, "; ")
}

func (ve *ValidationErrors) Add(field, message string) {
	if ve.Errors == nil {
		ve.Errors = make(map[string]string)
	}
	if _, exists := ve.Errors[field]; !exists {
		ve.Errors[field] = message
	}
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}
