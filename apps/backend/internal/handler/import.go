package handler

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model/imports"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type ImportHandler struct {
	Handler
	service *service.ImportService
}

func NewImportHandler(s *server.Server, svc *service.ImportService) *ImportHandler {
	return &ImportHandler{
		Handler: NewHandler(s),
		service: svc,
	}
}

func (h *ImportHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *imports.ImportCreateRequest) (*imports.ImportResponse, error) {
			userID := GetUserID(c)

			file, err := c.FormFile("file")
			if err != nil {
				return nil, ErrInvalidID
			}

			f, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer f.Close()

			reader := csv.NewReader(f)
			records, err := reader.ReadAll()
			if err != nil {
				return nil, err
			}

			if len(records) < 2 {
				return nil, ErrInvalidID
			}

			totalRows := len(records) - 1

			imp, err := h.service.CreateImport(c.Request.Context(), userID, totalRows)
			if err != nil {
				return nil, err
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

			return &imports.ImportResponse{
				ID:          imp.ID.String(),
				Status:      string(imp.Status),
				TotalRows:   imp.TotalRows,
				SuccessRows: imp.SuccessRows,
				CreatedAt:   imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}, nil
		},
		http.StatusAccepted,
		&imports.ImportCreateRequest{},
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
				ID:          imp.ID.String(),
				Status:      string(imp.Status),
				TotalRows:   imp.TotalRows,
				SuccessRows: imp.SuccessRows,
				FailedRows:  failedRows,
				CreatedAt:   imp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}, nil
		},
		http.StatusOK,
		&imports.GetImportRequest{},
	)(c)
}
