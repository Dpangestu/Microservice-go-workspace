package http

import (
	"encoding/json"
	"net/http"

	"bkc_microservice/shared/validation"
)

// ValidationMiddleware validates request body
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate on POST, PUT
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Parse content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "content type must be application/json", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// StructValidator for generic struct validation
type StructValidator struct {
	validator *validation.Validator
}

func NewStructValidator() *StructValidator {
	return &StructValidator{
		validator: validation.NewValidator(),
	}
}

// ValidateUserCreate validates CreateUserRequest
func (sv *StructValidator) ValidateUserCreate(username, email, password string) *validation.ValidationErrors {
	errs := &validation.ValidationErrors{Errors: make(map[string]string)}

	if err := sv.validator.ValidateUsername(username); err != nil {
		errs.Add("username", err.Error())
	}

	if err := sv.validator.ValidateEmail(email); err != nil {
		errs.Add("email", err.Error())
	}

	if err := sv.validator.ValidatePassword(password); err != nil {
		errs.Add("password", err.Error())
	}

	return errs
}

// ResponseValidationError mengembalikan error dalam format konsisten
func ResponseValidationError(w http.ResponseWriter, errs *validation.ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	resp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "VALIDATION_ERROR",
			"message": "Request validation failed",
			"details": errs.Errors,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
