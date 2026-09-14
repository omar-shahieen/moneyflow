package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/sqlerr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GlobalMiddlewares struct {
	server *server.Server
}

func NewGlobalMiddlewares(s *server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: s,
	}
}

// CORS mirrors the Echo CORSWithConfig middleware, allowing the origins
// configured in server.Config.Server.CORSAllowedOrigins.
func (global *GlobalMiddlewares) CORS() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = global.server.Config.Server.CORSAllowedOrigins
	return cors.New(cfg)
}

// RequestLogger logs one structured line per request. It must be
// registered so that it wraps ErrorHandler (see package doc above) — by
// the time it reads c.Writer.Status(), the final status is already set.
func (global *GlobalMiddlewares) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		req := c.Request

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		// Get enhanced logger from context
		logger := GetLogger(c)

		var e *zerolog.Event

		switch {
		case status >= 500:
			var reqErr error
			if len(c.Errors) > 0 {
				reqErr = c.Errors.Last().Err
			}
			e = logger.Error().Err(reqErr)
		case status >= 400:
			e = logger.Warn()
		default:
			e = logger.Info()
		}

		// Add request ID if available
		if requestID := GetRequestID(c); requestID != "" {
			e = e.Str("request_id", requestID)
		}

		// Add user context if available
		if userID := GetUserID(c); userID != "" {
			e = e.Str("user_id", userID)
		}

		e.
			Dur("latency", latency).
			Int("status", status).
			Str("method", req.Method).
			Str("uri", req.RequestURI).
			Str("host", req.Host).
			Str("ip", c.ClientIP()).
			Str("user_agent", req.UserAgent()).
			Msg("API")
	}
}

// Recover catches panics and routes them through the same error-handling
// logic as ErrorHandler, instead of letting Gin's default recovery just
// dump a bare 500.
func (global *GlobalMiddlewares) Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		err, ok := recovered.(error)
		if !ok {
			err = fmt.Errorf("%v", recovered)
		}
		global.handleError(err, c)
		c.Abort()
	})
}

// Secure sets the same default security headers as Echo's
// middleware.Secure() (no HSTS or CSP unless you add them here).
func (global *GlobalMiddlewares) Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Next()
	}
}

// ErrorHandler is the Gin equivalent of Echo's GlobalErrorHandler /
// HTTPErrorHandler. Handlers that need to fail should call c.Error(err)
// (and return); this middleware picks that error up after c.Next() and
// writes the JSON error response.
func (global *GlobalMiddlewares) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		global.handleError(c.Errors.Last().Err, c)
	}
}

// NoRouteHandler mirrors Echo's automatic conversion of a 404 into
// errs.NewNotFoundError. Register with router.NoRoute(mw.NoRouteHandler).
func (global *GlobalMiddlewares) NoRouteHandler(c *gin.Context) {
	global.handleError(errs.NewNotFoundError("Route not found", false, nil), c)
}

// handleError contains the shared conversion-and-response logic used by
// both ErrorHandler and Recover, equivalent to Echo's GlobalErrorHandler.
func (global *GlobalMiddlewares) handleError(err error, c *gin.Context) {
	originalErr := err

	// Try to handle known database errors, converting them to
	// application errors — only if not already an *errs.HTTPError.
	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		err = sqlerr.HandleError(err)
	}

	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	if errors.As(err, &httpErr) {
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action
	} else {
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusInternalServerError))
		message = http.StatusText(http.StatusInternalServerError)
	}

	// Log the original error to help with debugging. Use the enhanced
	// logger from context, which already includes request_id, method,
	// path, ip, user context, and trace context.
	logger := *GetLogger(c)

	logger.Error().Stack().
		Err(originalErr).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	if !c.Writer.Written() {
		c.JSON(status, errs.HTTPError{
			Code:     code,
			Message:  message,
			Status:   status,
			Override: httpErr != nil && httpErr.Override,
			Errors:   fieldErrors,
			Action:   action,
		})
	}
}
