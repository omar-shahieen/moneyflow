package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/handler"
)

func registerUserRoutes(r *gin.RouterGroup, h *handler.UserHandler) {
	user := r.Group("/user")
	{
		user.GET("", h.GetProfile)
		user.PATCH("", h.UpdateProfile)

		notifications := user.Group("/notifications")
		{
			notifications.GET("", h.ListNotifications)
			notifications.POST("", h.CreateNotification)
			notifications.GET("/unread-count", h.GetUnreadCount)
			notifications.PATCH("/read-all", h.MarkAllNotificationsAsRead)
			notifications.GET("/:id", h.GetNotification)
			notifications.PATCH("/:id/read", h.MarkNotificationAsRead)
			notifications.DELETE("/:id", h.DeleteNotification)
		}
	}
}
