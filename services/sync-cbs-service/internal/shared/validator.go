package shared

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(data interface{}) error {
	err := validate.Struct(data)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages string
			for _, fieldErr := range validationErrors {
				messages += fmt.Sprintf("field '%s' failed validation: %s; ", fieldErr.Field(), fieldErr.Tag())
			}
			return fmt.Errorf("validation error: %s", messages)
		}
		return err
	}
	return nil
}

func ValidateVar(data interface{}, tag string) error {
	return validate.Var(data, tag)
}
