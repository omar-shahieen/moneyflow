package middleware

import (
	"net/http"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/errs"
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
		// Define Clerk's failure handler
		failureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			auth.server.Logger.Error().
				Str("function", "RequireAuth").
				Dur("duration", time.Since(start)).
				Msg("could not get session claims from context")

			// Use Gin's AbortWithStatusJSON instead of manual encoding
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":     "UNAUTHORIZED",
				"message":  "Unauthorized",
				"override": "false",
				"status":   "401",
			})
		})

		// Initialize Clerk middleware with the custom failure handler
		clerkMiddleware := clerkhttp.WithHeaderAuthorization(
			clerkhttp.AuthorizationFailureHandler(failureHandler),
		)

		// Define the next step that executes if Clerk authentication succeeds
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Clerk's middleware attaches claims to the *http.Request context.
			// We must update Gin's request to preserve this context downstream.
			c.Request = r

			claims, ok := clerk.SessionClaimsFromContext(r.Context())
			if !ok {
				auth.server.Logger.Error().
					Str("function", "RequireAuth").
					Str("request_id", GetRequestID(c)). // Assuming GetRequestID accepts *gin.Context
					Dur("duration", time.Since(start)).
					Msg("could not get session claims from context")

				// Forward the error to Gin's error handling and stop execution
				c.Error(errs.NewUnauthorizedError("Unauthorized", false))
				c.Abort()
				return
			}

			// Set the claims in the Gin context
			c.Set("user_id", claims.Subject)
			c.Set("user_role", claims.ActiveOrganizationRole)
			c.Set("permissions", claims.Claims.ActiveOrganizationPermissions)

			auth.server.Logger.Info().
				Str("function", "RequireAuth").
				Str("user_id", claims.Subject).
				Str("request_id", GetRequestID(c)).
				Dur("duration", time.Since(start)).
				Msg("user authenticated successfully")

			// Proceed to the actual Gin handler
			c.Next()
		})

		// Execute the standard HTTP middleware chain
		clerkMiddleware(nextHandler).ServeHTTP(c.Writer, c.Request)
	}
}
