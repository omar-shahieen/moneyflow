package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model/transaction"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type ReceiptHandler struct {
	Handler
	storage ports.Storage
}

func NewReceiptHandler(s *server.Server, storage ports.Storage) *ReceiptHandler {
	return &ReceiptHandler{
		Handler: NewHandler(s),
		storage: storage,
	}
}

type PresignedUploadResponse struct {
	UploadURL  string    `json:"upload_url"`
	StorageKey string    `json:"storage_key"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (h *ReceiptHandler) GetUploadURL(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *transaction.GetReceiptUploadURLRequest) (*PresignedUploadResponse, error) {
			userID := GetUserID(c)

			storageKey := fmt.Sprintf("receipts/%s/%s.jpg", userID, req.ID.String())
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				contentType = "image/jpeg"
			}

			uploadURL, err := h.storage.GenerateUploadURL(storageKey, contentType, 15*time.Minute)
			if err != nil {
				return nil, err
			}

			return &PresignedUploadResponse{
				UploadURL:  uploadURL,
				StorageKey: storageKey,
				ExpiresAt:  time.Now().Add(15 * time.Minute),
			}, nil
		},
		http.StatusOK,
		&transaction.GetReceiptUploadURLRequest{},
	)(c)
}
