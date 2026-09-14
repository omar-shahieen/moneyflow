package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/report"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type ReportHandler struct {
	Handler
	reportService *service.ReportService
}

func NewReportHandler(s *server.Server, reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{
		Handler:       NewHandler(s),
		reportService: reportService,
	}
}

type ReportResponse struct {
	ID          string  `json:"id"`
	Format      string  `json:"format"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	Status      string  `json:"status"`
	DownloadURL *string `json:"download_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

func toReportResponse(r *report.Report, downloadURL string) ReportResponse {
	resp := ReportResponse{
		ID:          r.ID.String(),
		Format:      string(r.Format),
		PeriodStart: r.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   r.PeriodEnd.Format("2006-01-02"),
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if r.CompletedAt != nil {
		s := r.CompletedAt.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &s
	}
	if r.Status == report.ReportStatusReady && downloadURL != "" {
		resp.DownloadURL = &downloadURL
	}
	return resp
}

func (h *ReportHandler) List(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, query *report.ListReportsRequest) (*model.PaginatedResponse[ReportResponse], error) {
			userID := GetUserID(c)

			result, err := h.reportService.GetReports(c.Request.Context(), userID, query)
			if err != nil {
				return nil, err
			}

			responses := make([]ReportResponse, 0, len(result.Data))
			for _, r := range result.Data {
				var downloadURL string
				if r.Status == report.ReportStatusReady {
					url, err := h.reportService.GetDownloadURL(c.Request.Context(), r.ID, userID)
					if err == nil {
						downloadURL = url
					}
				}
				responses = append(responses, toReportResponse(&r, downloadURL))
			}

			return &model.PaginatedResponse[ReportResponse]{
				Data:       responses,
				Page:       result.Page,
				Limit:      result.Limit,
				Total:      result.Total,
				TotalPages: result.TotalPages,
			}, nil
		},
		http.StatusOK,
		&report.ListReportsRequest{},
	)(c)
}

func (h *ReportHandler) GetByID(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *report.GetReportRequest) (*ReportResponse, error) {
			userID := GetUserID(c)

			r, err := h.reportService.GetReportByID(c.Request.Context(), userID, req.ID)
			if err != nil {
				return nil, err
			}

			var downloadURL string
			if r.Status == report.ReportStatusReady {
				url, err := h.reportService.GetDownloadURL(c.Request.Context(), req.ID, userID)
				if err == nil {
					downloadURL = url
				}
			}

			resp := toReportResponse(r, downloadURL)
			return &resp, nil
		},
		http.StatusOK,
		&report.GetReportRequest{},
	)(c)
}

func (h *ReportHandler) Create(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *report.CreateReportRequest) (*ReportResponse, error) {
			userID := GetUserID(c)

			r, err := h.reportService.CreateReport(c.Request.Context(), userID, payload)
			if err != nil {
				return nil, err
			}

			resp := toReportResponse(r, "")
			return &resp, nil
		},
		http.StatusAccepted,
		&report.CreateReportRequest{},
	)(c)
}
