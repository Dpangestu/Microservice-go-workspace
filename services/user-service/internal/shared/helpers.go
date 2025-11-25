package shared

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PtrString returns pointer to string
func PtrString(s string) *string {
	return &s
}

// PtrInt returns pointer to int
func PtrInt(i int) *int {
	return &i
}

// PtrBool returns pointer to bool
func PtrBool(b bool) *bool {
	return &b
}

// StringValue returns value from string pointer
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// IntValue returns value from int pointer
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// BoolValue returns value from bool pointer
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// WrapError wraps error dengan konteks tambahan
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// GenerateUUID menghasilkan UUID v4 (random)
func GenerateUUID() string {
	return uuid.New().String()
}

// GetNowTime mengembalikan waktu saat ini (UTC)
func GetNowTime() time.Time {
	return time.Now().UTC()
}
