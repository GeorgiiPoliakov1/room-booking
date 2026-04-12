// internal/dto/validator/errors.go
package validator

import (
	"github.com/go-playground/validator"
)

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if err == nil {
		return errors
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
				Value:   e.Value(),
			})
		}
	}

	return errors
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "minimum " + e.Param() + " characters required"
	case "max":
		return "maximum " + e.Param() + " characters allowed"
	case "room_name":
		return "room name can only contain letters, numbers, spaces, hyphens and underscores"
	case "role":
		return "role must be either 'admin' or 'user'"
	case "time_slot":
		return "time must be in HH:MM format (24-hour)"
	case "gtfield":
		return "must be greater than " + e.Param()
	case "ltfield":
		return "must be less than " + e.Param()
	case "oneof":
		return "must be one of: " + e.Param()
	case "datetime":
		return "invalid time format"
	default:
		return "invalid value for " + e.Tag()
	}
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(validator.ValidationErrors)
	return ok
}
