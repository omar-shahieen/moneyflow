package handler

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type ImportHandler struct {
	service *service.ImportService
}

func NewImportHandler(s *service.ImportService) *ImportHandler {
	return &ImportHandler{service: s}
}

type ImportResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	TotalRows   int                    `json:"total_rows"`
	SuccessRows int                    `json:"success_rows"`
	FailedRows  []model.ImportRowError `json:"failed_rows"`
	CreatedAt   string                 `json:"created_at"`
}

func (h *ImportHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	f, err := file.Open()
	if err != nil {
		RespondError(c, err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		RespondError(c, err)
		return
	}

	if len(records) < 2 {
		RespondError(c, ErrInvalidID)
		return
	}

	totalRows := len(records) - 1

	imp, err := h.service.Create(c.Request.Context(), userID, service.CreateImportInput{
		TotalRows: totalRows,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	go func() {
		var rows []service.CSVRow
		for _, record := range records[1:] {
			if len(record) < 4 {
				continue
			}

			amount, _ := strconv.ParseInt(record[1], 10, 64)

			rows = append(rows, service.CSVRow{
				CategoryName: record[0],
				Amount:       amount,
				Currency:     record[2],
				Note:         record[3],
				Date: func() string {
					if len(record) > 4 {
						return record[4]
					}
					return ""
				}(),
			})
		}
		_ = h.service.ProcessImport(c.Request.Context(), imp.ID, userID, rows)
	}()

	resp := ImportResponse{
		ID:          imp.ID.String(),
		Status:      string(imp.Status),
		TotalRows:   imp.TotalRows,
		SuccessRows: imp.SuccessRows,
		CreatedAt:   imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	RespondJSON(c, http.StatusAccepted, resp)
}

func (h *ImportHandler) GetByID(c *gin.Context) {
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

	imp, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	var failedRows []model.ImportRowError
	_ = json.Unmarshal(imp.FailedRows, &failedRows)
	if failedRows == nil {
		failedRows = []model.ImportRowError{}
	}

	RespondJSON(c, http.StatusOK, ImportResponse{
		ID:          imp.ID.String(),
		Status:      string(imp.Status),
		TotalRows:   imp.TotalRows,
		SuccessRows: imp.SuccessRows,
		FailedRows:  failedRows,
		CreatedAt:   imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ImportHandler) GetUploadURL(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "text/csv") && !strings.Contains(contentType, "application/csv") {
		RespondError(c, ErrInvalidID)
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"message": "Use POST /api/v1/imports with multipart/form-data",
	})
}
