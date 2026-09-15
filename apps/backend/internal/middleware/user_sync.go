package middleware

import (
	"errors"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	usermodel "github.com/omar-shahieen/moneyflow/internal/model/user"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type UserSyncMiddleware struct {
	server   *server.Server
	userRepo *repository.UserRepo
}

func NewUserSyncMiddleware(s *server.Server) *UserSyncMiddleware {
	return &UserSyncMiddleware{
		server:   s,
		userRepo: repository.NewUserRepository(s),
	}
}

func (m *UserSyncMiddleware) SyncUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}

		_, err := m.userRepo.GetByID(c.Request.Context(), userID)
		if err == nil {
			c.Next()
			return
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			m.server.Logger.Error().
				Str("function", "SyncUser").
				Str("user_id", userID).
				Err(err).
				Msg("failed to check user existence")
			c.Next()
			return
		}

		clerkUser, err := clerkuser.Get(c.Request.Context(), userID)
		if err != nil {
			m.server.Logger.Error().
				Str("function", "SyncUser").
				Str("user_id", userID).
				Err(err).
				Msg("failed to fetch user from Clerk")
			c.Next()
			return
		}

		newUser := &usermodel.UserAccount{
			ID:          clerkUser.ID,
			Email:       extractEmail(clerkUser),
			DisplayName: buildDisplayName(clerkUser),
			AvatarURL:   derefString(clerkUser.ImageURL),
			Timezone:    "Africa/Cairo",
		}

		if err := m.userRepo.Create(c.Request.Context(), newUser); err != nil {
			m.server.Logger.Error().
				Str("function", "SyncUser").
				Str("user_id", userID).
				Err(err).
				Msg("failed to create user in database")
			c.Next()
			return
		}

		m.server.Logger.Info().
			Str("function", "SyncUser").
			Str("user_id", userID).
			Msg("user synced successfully")

		c.Next()
	}
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func extractEmail(u *clerk.User) string {
	primaryID := derefString(u.PrimaryEmailAddressID)
	if primaryID != "" && u.EmailAddresses != nil {
		for _, email := range u.EmailAddresses {
			if email.ID == primaryID {
				return email.EmailAddress
			}
		}
	}

	if len(u.EmailAddresses) > 0 {
		return u.EmailAddresses[0].EmailAddress
	}

	return ""
}

func buildDisplayName(u *clerk.User) string {
	firstName := derefString(u.FirstName)
	lastName := derefString(u.LastName)
	username := derefString(u.Username)

	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	}
	if firstName != "" {
		return firstName
	}
	if lastName != "" {
		return lastName
	}
	return username
}
