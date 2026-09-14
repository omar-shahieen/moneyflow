package router

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
	"github.com/omar-shahieen/moneyflow/internal/middleware"
	v1 "github.com/omar-shahieen/moneyflow/internal/router/v1"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"golang.org/x/time/rate"
)

func NewRouter(s *server.Server, h *handler.Handlers) *gin.Engine {
	middlewares := middleware.NewMiddlewares(s)

	router := gin.New()

	rateLimiter := func() gin.HandlerFunc {
		var ips = make(map[string]*rate.Limiter)
		var mu sync.Mutex

		return func(c *gin.Context) {
			ip := c.ClientIP()

			mu.Lock()
			limiter, exists := ips[ip]
			if !exists {
				limiter = rate.NewLimiter(rate.Limit(20), 1)
				ips[ip] = limiter
			}
			mu.Unlock()

			if !limiter.Allow() {
				if rateLimitMiddleware := middlewares.RateLimit; rateLimitMiddleware != nil {
					rateLimitMiddleware.RecordRateLimitHit(c.Request.URL.Path)
				}

				s.Logger.Warn().
					Str("request_id", middleware.GetRequestID(c)).
					Str("identifier", ip).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("ip", ip).
					Msg("rate limit exceeded")

				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Rate limit exceeded"})
				return
			}
			c.Next()
		}
	}()

	router.Use(
		rateLimiter,
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		middleware.RequestID(),
		middlewares.Tracing.NewRelicMiddleware(),
		middlewares.Tracing.EnhanceTracing(),
		middlewares.ContextEnhancer.EnhanceContext(),
		middlewares.Global.RequestLogger(),
		middlewares.Global.Recover(),
	)

	registerSystemRoutes(router, h)

	v1Group := router.Group("/api/v1")
	v1Group.Use(middlewares.Auth.RequireAuth())

	v1.RegisterV1Routes(v1Group, h)

	return router
}
