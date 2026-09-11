package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/domain"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/sqlerr"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code     string           `json:"code"`
	Message  string           `json:"message"`
	Status   int              `json:"status"`
	Override bool             `json:"override,omitempty"`
	Errors   []fieldErrorBody `json:"errors,omitempty"`
}

type fieldErrorBody struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleGinError(c, err)
		}
	}
}

func handleGinError(c *gin.Context, err error) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		c.JSON(domainErr.Status, errorResponse{
			Error: errorBody{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Status:  domainErr.Status,
			},
		})
		return
	}

	var httpErr *errs.HTTPError
	if errors.As(err, &httpErr) {
		fieldErrors := make([]fieldErrorBody, 0, len(httpErr.Errors))
		for _, fe := range httpErr.Errors {
			fieldErrors = append(fieldErrors, fieldErrorBody{
				Field: fe.Field,
				Error: fe.Error,
			})
		}

		c.JSON(httpErr.Status, errorResponse{
			Error: errorBody{
				Code:     httpErr.Code,
				Message:  httpErr.Message,
				Status:   httpErr.Status,
				Override: httpErr.Override,
				Errors:   fieldErrors,
			},
		})
		return
	}

	dbErr := sqlerr.HandleError(err)
	var sqlHttpErr *errs.HTTPError
	if errors.As(dbErr, &sqlHttpErr) {
		c.JSON(sqlHttpErr.Status, errorResponse{
			Error: errorBody{
				Code:    sqlHttpErr.Code,
				Message: sqlHttpErr.Message,
				Status:  sqlHttpErr.Status,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
			Status:  http.StatusInternalServerError,
		},
	})
}
