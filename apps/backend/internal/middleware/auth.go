package middleware

import (
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{
		server: s,
	}
}

func (auth *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Unauthorized",
				"status":  "401",
			})
			return
		}

		failureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			auth.server.Logger.Error().
				Str("function", "RequireAuth").
				Dur("duration", time.Since(start)).
				Msg("could not get session claims from context")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Unauthorized",
				"status":  "401",
			})
		})

		clerkMiddleware := clerkhttp.WithHeaderAuthorization(
			clerkhttp.AuthorizationFailureHandler(failureHandler),
		)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			c.Request = r

			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				auth.server.Logger.Error().
					Str("function", "RequireAuth").
					Str("request_id", GetRequestID(c)).
					Dur("duration", time.Since(start)).
					Msg("could not get session claims from context")

				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Unauthorized",
					"status":  "401",
				})
				return
			}

			c.Set("user_id", claims.Subject)
			c.Set("user_role", claims.ActiveOrganizationRole)
			c.Set("permissions", claims.Claims.ActiveOrganizationPermissions)

			auth.server.Logger.Info().
				Str("function", "RequireAuth").
				Str("user_id", claims.Subject).
				Str("request_id", GetRequestID(c)).
				Dur("duration", time.Since(start)).
				Msg("user authenticated successfully")

			c.Next()
		})

		clerkMiddleware(nextHandler).ServeHTTP(c.Writer, c.Request)
	}
}
