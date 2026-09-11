package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	GinKeyRequestID = "request_id"
	GinKeyUserID    = "user_id"
	GinKeyLogger    = "logger"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(GinKeyRequestID, requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

func GetRequestIDFromGin(c *gin.Context) string {
	if id, exists := c.Get(GinKeyRequestID); exists {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	return ""
}

func GetUserIDFromGin(c *gin.Context) string {
	if id, exists := c.Get(GinKeyUserID); exists {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	return ""
}

func LoggerMiddleware(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID := GetRequestIDFromGin(c)
		userID := GetUserIDFromGin(c)

		var logEvent *zerolog.Event
		switch {
		case status >= 500:
			logEvent = log.Error()
		case status >= 400:
			logEvent = log.Warn()
		default:
			logEvent = log.Info()
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("uri", c.Request.RequestURI).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent())

		if userID != "" {
			logEvent.Str("user_id", userID)
		}

		if len(c.Errors) > 0 {
			logEvent.Str("errors", c.Errors.String())
		}

		logEvent.Msg("API")
	}
}

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func SecureMiddleware() gin.HandlerFunc {
	return secure.New(secure.Config{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	})
}

func RecoveryMiddleware(log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestIDFromGin(c)

				log.Error().
					Interface("error", err).
					Str("request_id", requestID).
					Str("method", c.Request.Method).
					Str("uri", c.Request.RequestURI).
					Msg("panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An internal error occurred",
						"status":  500,
					},
				})
			}
		}()
		c.Next()
	}
}

func GinRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func ContextEnrichmentMiddleware(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestIDFromGin(c)

		contextLogger := log.With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Str("ip", c.ClientIP()).
			Logger()

		c.Set(GinKeyLogger, &contextLogger)
		c.Next()
	}
}

func GetLoggerFromGin(c *gin.Context) *zerolog.Logger {
	if logger, exists := c.Get(GinKeyLogger); exists {
		if log, ok := logger.(*zerolog.Logger); ok {
			return log
		}
	}
	nop := zerolog.Nop()
	return &nop
}

func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header required",
					"status":  401,
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization format",
					"status":  401,
				},
			})
			return
		}

		c.Next()
	}
}
