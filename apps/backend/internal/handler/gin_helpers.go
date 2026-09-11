package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/middleware"
	"github.com/omar-shahieen/moneyflow/internal/sqlerr"
	"github.com/rs/zerolog"
)

type GinHandler struct {
	logger zerolog.Logger
}

func NewGinHandler(log zerolog.Logger) *GinHandler {
	return &GinHandler{logger: log}
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Status  int              `json:"status"`
	Errors  []FieldErrorBody `json:"errors,omitempty"`
}

type FieldErrorBody struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"pagination"`
}

func RespondJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, SuccessResponse{Data: data})
}

func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func RespondError(c *gin.Context, err error) {
	logger := middleware.GetLoggerFromGin(c)
	logger.Error().Err(err).Msg("request error")

	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		c.JSON(domainErr.Status, ErrorResponse{
			Error: ErrorBody{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Status:  domainErr.Status,
			},
		})
		return
	}

	var httpErr *errs.HTTPError
	if errors.As(err, &httpErr) {
		fieldErrors := make([]FieldErrorBody, 0, len(httpErr.Errors))
		for _, fe := range httpErr.Errors {
			fieldErrors = append(fieldErrors, FieldErrorBody{
				Field: fe.Field,
				Error: fe.Error,
			})
		}

		c.JSON(httpErr.Status, ErrorResponse{
			Error: ErrorBody{
				Code:    httpErr.Code,
				Message: httpErr.Message,
				Status:  httpErr.Status,
				Errors:  fieldErrors,
			},
		})
		return
	}

	dbErr := sqlerr.HandleError(err)
	var sqlHttpErr *errs.HTTPError
	if errors.As(dbErr, &sqlHttpErr) {
		c.JSON(sqlHttpErr.Status, ErrorResponse{
			Error: ErrorBody{
				Code:    sqlHttpErr.Code,
				Message: sqlHttpErr.Message,
				Status:  sqlHttpErr.Status,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: ErrorBody{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
			Status:  http.StatusInternalServerError,
		},
	})
}

func BindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		message := err.Error()
		if parts := strings.Split(message, "\n"); len(parts) > 0 {
			message = parts[0]
		}
		return errs.NewBadRequestError(message, false, nil, nil, nil)
	}
	return nil
}

func GetUserID(c *gin.Context) string {
	return middleware.GetUserIDFromGin(c)
}

func GetRequestID(c *gin.Context) string {
	return middleware.GetRequestIDFromGin(c)
}
