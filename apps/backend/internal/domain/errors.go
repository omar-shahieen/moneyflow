package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("resource conflict")
	ErrPlanLimitExceeded = errors.New("plan limit exceeded")
	ErrValidation        = errors.New("validation failed")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrRateLimited       = errors.New("rate limited")
	ErrInternal          = errors.New("internal error")
)

type DomainError struct {
	Err     error
	Code    string
	Status  int
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewDomainError(err error, code string, status int, message string) *DomainError {
	return &DomainError{
		Err:     err,
		Code:    code,
		Status:  status,
		Message: message,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsPlanLimitExceeded(err error) bool {
	return errors.Is(err, ErrPlanLimitExceeded)
}

func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}
