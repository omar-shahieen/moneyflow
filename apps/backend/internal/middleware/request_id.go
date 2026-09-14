package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String() // 4c90fc3f-39cc-4b04-af21-c83ee64aa67e
		}

		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestIDKey)
}
