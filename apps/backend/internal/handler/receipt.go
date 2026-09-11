package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/ports"
)

type ReceiptHandler struct {
	storage ports.Storage
}

func NewReceiptHandler(storage ports.Storage) *ReceiptHandler {
	return &ReceiptHandler{storage: storage}
}

type PresignedUploadResponse struct {
	UploadURL  string    `json:"upload_url"`
	StorageKey string    `json:"storage_key"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (h *ReceiptHandler) GetUploadURL(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	storageKey := fmt.Sprintf("receipts/%s/%s.jpg", userID, id.String())
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	uploadURL, err := h.storage.GenerateUploadURL(storageKey, contentType, 15*time.Minute)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusOK, PresignedUploadResponse{
		UploadURL:  uploadURL,
		StorageKey: storageKey,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
	})
}
