package middleware

import (
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type RateLimitMiddleware struct {
	server *server.Server
}

func NewRateLimitMiddleware(s *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		server: s,
	}
}

func (rl *RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	if rl.server.LoggerService != nil && rl.server.LoggerService.GetApplication() != nil {
		rl.server.LoggerService.GetApplication().RecordCustomEvent("RateLimitHit", map[string]interface{}{
			"endpoint": endpoint,
		})
	}
}
