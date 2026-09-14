package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/omar-shahieen/moneyflow/internal/logger"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/rs/zerolog"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

type ContextEnhancer struct {
	server *server.Server
}

func NewContextEnhancer(s *server.Server) *ContextEnhancer {
	return &ContextEnhancer{server: s}
}

func (ce *ContextEnhancer) EnhanceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract request ID
		// NOTE: assumes a gin-compatible GetRequestID(c *gin.Context); update the
		// call if your request-ID middleware is still Echo-specific.
		requestID := GetRequestID(c)

		// Create enhanced logger with request context
		contextLogger := ce.server.Logger.With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Str("ip", c.ClientIP()).
			Logger()

		// Add trace context if available
		if txn := newrelic.FromContext(c.Request.Context()); txn != nil {
			contextLogger = logger.WithTraceContext(contextLogger, txn)
		}

		// Extract user information from JWT token or session
		if userID := ce.extractUserID(c); userID != "" {
			contextLogger = contextLogger.With().Str("user_id", userID).Logger()
		}

		if userRole := ce.extractUserRole(c); userRole != "" {
			contextLogger = contextLogger.With().Str("user_role", userRole).Logger()
		}

		// Store the enhanced logger in context
		c.Set(LoggerKey, &contextLogger)

		// Create a new request context with the logger and swap it in
		ctx := context.WithValue(c.Request.Context(), LoggerKey, &contextLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (ce *ContextEnhancer) extractUserID(c *gin.Context) string {
	// Check if user_id was already set by auth middleware (Clerk).
	// GetString returns "" if the key is unset or not a string, so no extra
	// type-assertion boilerplate is needed here (unlike Echo's c.Get).
	return c.GetString(UserIDKey)
}

func (ce *ContextEnhancer) extractUserRole(c *gin.Context) string {
	// Check if user_role was set by auth middleware (Clerk)
	return c.GetString(UserRoleKey)
}

func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

func GetLogger(c *gin.Context) *zerolog.Logger {
	if val, ok := c.Get(LoggerKey); ok {
		if l, ok := val.(*zerolog.Logger); ok {
			return l
		}
	}
	// Fallback to a basic logger if not found
	l := zerolog.Nop()
	return &l
}
