package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type ImportHandler struct {
	Handler
	service *service.ImportService
	storage ports.Storage
}

func NewImportHandler(s *server.Server, svc *service.ImportService, storage ports.Storage) *ImportHandler {
	return &ImportHandler{
		Handler: NewHandler(s),
		service: svc,
		storage: storage,
	}
}

func (h *ImportHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *imports.ImportCreateRequest) (*imports.ImportCreateResponse, error) {
			userID := GetUserID(c)

			imp, err := h.service.CreateImport(c.Request.Context(), userID, req.TotalRows)
			if err != nil {
				return nil, err
			}

			presignExpiry := 15 * time.Minute
			if h.server.Config.Storage.R2.PresignExpiry > 0 {
				presignExpiry = time.Duration(h.server.Config.Storage.R2.PresignExpiry) * time.Minute
			}

			uploadURL, err := h.storage.GenerateUploadURL(imp.StorageKey, "text/csv", presignExpiry)
			if err != nil {
				return nil, err
			}

			return &imports.ImportCreateResponse{
				ImportID:  imp.ID.String(),
				UploadURL: uploadURL,
				ExpiresAt: time.Now().Add(presignExpiry),
			}, nil
		},
		http.StatusAccepted,
		&imports.ImportCreateRequest{},
	)(c)
}

func (h *ImportHandler) ConfirmUpload(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *imports.GetImportRequest) (*imports.ImportResponse, error) {
			userID := GetUserID(c)

			imp, err := h.service.ConfirmUpload(c.Request.Context(), req.ID, userID)
			if err != nil {
				return nil, err
			}

			var failedRows []imports.ImportRowError
			_ = json.Unmarshal(imp.FailedRows, &failedRows)
			if failedRows == nil {
				failedRows = []imports.ImportRowError{}
			}

			return &imports.ImportResponse{
				ID:             imp.ID.String(),
				Status:         string(imp.Status),
				TotalRows:      imp.TotalRows,
				SuccessRows:    imp.SuccessRows,
				FailedRows:     failedRows,
				CreatedAt:      imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
				StorageKey:     imp.StorageKey,
				ChecksumSHA256: imp.ChecksumSHA256,
			}, nil
		},
		http.StatusOK,
		&imports.GetImportRequest{},
	)(c)
}

func (h *ImportHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *imports.ListImportsRequest) (*model.PaginatedResponse[imports.ImportResponse], error) {
			userID := GetUserID(c)
			result, err := h.service.GetImports(c.Request.Context(), userID, req)
			if err != nil {
				return nil, err
			}

			responses := make([]imports.ImportResponse, len(result.Data))
			for i, imp := range result.Data {
				var failedRows []imports.ImportRowError
				_ = json.Unmarshal(imp.FailedRows, &failedRows)
				if failedRows == nil {
					failedRows = []imports.ImportRowError{}
				}
				responses[i] = imports.ImportResponse{
					ID:             imp.ID.String(),
					Status:         string(imp.Status),
					TotalRows:      imp.TotalRows,
					SuccessRows:    imp.SuccessRows,
					FailedRows:     failedRows,
					CreatedAt:      imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
					StorageKey:     imp.StorageKey,
					ChecksumSHA256: imp.ChecksumSHA256,
				}
			}

			return &model.PaginatedResponse[imports.ImportResponse]{
				Data:       responses,
				Page:       result.Page,
				Limit:      result.Limit,
				Total:      result.Total,
				TotalPages: result.TotalPages,
			}, nil
		},
		http.StatusOK,
		&imports.ListImportsRequest{},
	)(c)
}

func (h *ImportHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *imports.GetImportRequest) (*imports.ImportResponse, error) {
			userID := GetUserID(c)

			imp, err := h.service.GetImportByID(c.Request.Context(), userID, req.ID)
			if err != nil {
				return nil, err
			}

			var failedRows []imports.ImportRowError
			_ = json.Unmarshal(imp.FailedRows, &failedRows)
			if failedRows == nil {
				failedRows = []imports.ImportRowError{}
			}

			return &imports.ImportResponse{
				ID:             imp.ID.String(),
				Status:         string(imp.Status),
				TotalRows:      imp.TotalRows,
				SuccessRows:    imp.SuccessRows,
				FailedRows:     failedRows,
				CreatedAt:      imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
				StorageKey:     imp.StorageKey,
				ChecksumSHA256: imp.ChecksumSHA256,
			}, nil
		},
		http.StatusOK,
		&imports.GetImportRequest{},
	)(c)
}
