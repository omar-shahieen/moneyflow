package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var errUnexpected = errors.New("something went wrong")

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondJSON(c, http.StatusOK, gin.H{"message": "ok"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Error("expected data to be a map")
		return
	}
	if data["message"] != "ok" {
		t.Errorf("expected message 'ok', got '%v'", data["message"])
	}
}

func TestRespondError_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, ErrUnauthorized)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", resp.Error.Code)
	}
}

func TestRespondError_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, ErrInvalidID)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Error.Code != "INVALID_ID" {
		t.Errorf("expected error code INVALID_ID, got %s", resp.Error.Code)
	}
}

func TestRespondError_Internal(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, errUnexpected)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestBindAndValidate_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name": "Test Category"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := BindAndValidate(c, &input); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if input.Name != "Test Category" {
		t.Errorf("expected name 'Test Category', got '%s'", input.Name)
	}
}

func TestBindAndValidate_MissingField(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	err := BindAndValidate(c, &input)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestGetUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user_123")

	userID := GetUserID(c)
	if userID != "user_123" {
		t.Errorf("expected user_id 'user_123', got '%s'", userID)
	}
}

func TestGetRequestID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("request_id", "req_456")

	requestID := GetRequestID(c)
	if requestID != "req_456" {
		t.Errorf("expected request_id 'req_456', got '%s'", requestID)
	}
}
