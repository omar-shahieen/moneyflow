package domain

import (
	"testing"
)

func TestDomainError(t *testing.T) {
	err := NewDomainError(ErrNotFound, "TEST_ERROR", 404, "test message")

	if err.Code != "TEST_ERROR" {
		t.Errorf("expected code TEST_ERROR, got %s", err.Code)
	}
	if err.Status != 404 {
		t.Errorf("expected status 404, got %d", err.Status)
	}
	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got %s", err.Message)
	}
}

func TestDomainError_Error(t *testing.T) {
	err := NewDomainError(ErrNotFound, "TEST_ERROR", 404, "test message")

	if err.Error() != "test message" {
		t.Errorf("expected Error() to return 'test message', got %s", err.Error())
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	err := NewDomainError(ErrNotFound, "TEST_ERROR", 404, "test message")

	if err.Unwrap() != ErrNotFound {
		t.Error("expected Unwrap() to return ErrNotFound")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrNotFound,
		ErrForbidden,
		ErrConflict,
		ErrPlanLimitExceeded,
		ErrValidation,
		ErrUnauthorized,
		ErrRateLimited,
		ErrInternal,
	}

	for _, sentinel := range sentinels {
		if sentinel == nil {
			t.Error("sentinel error should not be nil")
		}
		if sentinel.Error() == "" {
			t.Error("sentinel error should have a message")
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	err := NewDomainError(ErrNotFound, "TEST_ERROR", 404, "test message")

	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true for DomainError wrapping ErrNotFound")
	}

	if IsForbidden(err) {
		t.Error("expected IsForbidden to return false for DomainError wrapping ErrNotFound")
	}
}
