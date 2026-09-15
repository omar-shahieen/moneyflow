package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validatable interface {
	Validate() error
}

// Normalizable is implemented by request DTOs that need to canonicalize input
// before it is validated and passed to the application layer.
type Normalizable interface {
	Normalize()
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

type ValidationError struct {
	Message     string
	FieldErrors []FieldError
}

type FieldError struct {
	Field string
	Error string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidateStruct(v Validatable) (string, []FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []FieldError) {
	var fieldErrors []FieldError
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		customValidationErrors, ok := err.(CustomValidationErrors)
		if ok {
			for _, e := range customValidationErrors {
				fieldErrors = append(fieldErrors, FieldError{
					Field: e.Field,
					Error: e.Message,
				})
			}
			return "Validation failed", fieldErrors
		}
		return err.Error(), nil
	}

	for _, e := range validationErrors {
		field := strings.ToLower(e.Field())
		var msg string

		switch e.Tag() {
		case "required":
			msg = "is required"
		case "min":
			if e.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must be at least %s characters", e.Param())
			} else {
				msg = fmt.Sprintf("must be at least %s", e.Param())
			}
		case "max":
			if e.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must not exceed %s characters", e.Param())
			} else {
				msg = fmt.Sprintf("must not exceed %s", e.Param())
			}
		case "oneof":
			msg = fmt.Sprintf("must be one of: %s", e.Param())
		case "email":
			msg = "must be a valid email address"
		case "e164":
			msg = "must be a valid phone number with country code"
		case "uuid":
			msg = "must be a valid UUID"
		case "uuidList":
			msg = "must be a comma-separated list of valid UUIDs"
		case "dive":
			msg = "some items are invalid"
		default:
			if e.Param() != "" {
				msg = fmt.Sprintf("%s: %s:%s", field, e.Tag(), e.Param())
			} else {
				msg = fmt.Sprintf("%s: %s", field, e.Tag())
			}
		}

		fieldErrors = append(fieldErrors, FieldError{
			Field: strings.ToLower(e.Field()),
			Error: msg,
		})
	}

	return "Validation failed", fieldErrors
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
