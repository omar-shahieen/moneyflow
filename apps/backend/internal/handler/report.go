package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/domain/model"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(s *service.ReportService) *ReportHandler {
	return &ReportHandler{service: s}
}

type CreateReportRequest struct {
	Format      string `json:"format" binding:"required"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
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

func (h *ReportHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	reports, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	if reports == nil {
		reports = []model.Report{}
	}

	var responses []ReportResponse
	for _, r := range reports {
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
		if r.Status == model.ReportStatusReady {
			url, err := h.service.GetDownloadURL(c.Request.Context(), r.ID, userID)
			if err == nil {
				resp.DownloadURL = &url
			}
		}
		responses = append(responses, resp)
	}

	RespondJSON(c, http.StatusOK, responses)
}

func (h *ReportHandler) GetByID(c *gin.Context) {
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

	report, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	resp := ReportResponse{
		ID:          report.ID.String(),
		Format:      string(report.Format),
		PeriodStart: report.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   report.PeriodEnd.Format("2006-01-02"),
		Status:      string(report.Status),
		CreatedAt:   report.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if report.CompletedAt != nil {
		s := report.CompletedAt.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &s
	}

	if report.Status == model.ReportStatusReady {
		url, err := h.service.GetDownloadURL(c.Request.Context(), id, userID)
		if err == nil {
			resp.DownloadURL = &url
		}
	}

	RespondJSON(c, http.StatusOK, resp)
}

func (h *ReportHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		RespondError(c, ErrUnauthorized)
		return
	}

	var req CreateReportRequest
	if err := BindAndValidate(c, &req); err != nil {
		RespondError(c, err)
		return
	}

	periodStart, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	periodEnd, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		RespondError(c, ErrInvalidID)
		return
	}

	report, err := h.service.Create(c.Request.Context(), userID, service.CreateReportInput{
		Format:      model.ReportFormat(req.Format),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondJSON(c, http.StatusAccepted, ReportResponse{
		ID:          report.ID.String(),
		Format:      string(report.Format),
		PeriodStart: report.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   report.PeriodEnd.Format("2006-01-02"),
		Status:      string(report.Status),
		CreatedAt:   report.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
