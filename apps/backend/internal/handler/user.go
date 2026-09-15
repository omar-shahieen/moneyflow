package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/notification"
	"github.com/omar-shahieen/moneyflow/internal/model/user"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
)

type UserHandler struct {
	Handler
	userService *service.UserService
}

func NewUserHandler(s *server.Server, userService *service.UserService) *UserHandler {
	return &UserHandler{
		Handler:     NewHandler(s),
		userService: userService,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) (*user.UserAccount, error) {
			userID := GetUserID(c)
			return h.userService.GetProfile(c.Request.Context(), userID)
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	Handle[*user.UpdateProfileRequest, *user.UserAccount](
		h.Handler,
		func(c *gin.Context, payload *user.UpdateProfileRequest) (*user.UserAccount, error) {
			userID := GetUserID(c)
			return h.userService.UpdateProfile(c.Request.Context(), userID, payload)
		},
		http.StatusOK,
		&user.UpdateProfileRequest{},
	)(c)
}

func (h *UserHandler) ListNotifications(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *notification.ListNotificationsRequest) (*model.PaginatedResponse[notification.Notification], error) {
			userID := GetUserID(c)
			return h.userService.GetNotifications(c.Request.Context(), userID, req)
		},
		http.StatusOK,
		&notification.ListNotificationsRequest{},
	)(c)
}

func (h *UserHandler) GetNotification(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *notification.GetNotificationRequest) (*notification.Notification, error) {
			userID := GetUserID(c)
			return h.userService.GetNotificationByID(c.Request.Context(), userID, req.ID)
		},
		http.StatusOK,
		&notification.GetNotificationRequest{},
	)(c)
}

func (h *UserHandler) CreateNotification(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, payload *notification.CreateNotificationRequest) (*notification.Notification, error) {
			userID := GetUserID(c)
			return h.userService.CreateNotification(c.Request.Context(), userID, payload)
		},
		http.StatusCreated,
		&notification.CreateNotificationRequest{},
	)(c)
}

func (h *UserHandler) MarkNotificationAsRead(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *notification.MarkAsReadRequest) error {
			userID := GetUserID(c)
			return h.userService.MarkNotificationAsRead(c.Request.Context(), userID, req.ID)
		},
		http.StatusNoContent,
		&notification.MarkAsReadRequest{},
	)(c)
}

func (h *UserHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *notification.MarkAllAsReadRequest) error {
			userID := GetUserID(c)
			return h.userService.MarkAllNotificationsAsRead(c.Request.Context(), userID)
		},
		http.StatusNoContent,
		&notification.MarkAllAsReadRequest{},
	)(c)
}

func (h *UserHandler) DeleteNotification(c *gin.Context) {
	HandleNoContent(
		h.Handler,
		func(c *gin.Context, req *notification.DeleteNotificationRequest) error {
			userID := GetUserID(c)
			return h.userService.DeleteNotification(c.Request.Context(), userID, req.ID)
		},
		http.StatusNoContent,
		&notification.DeleteNotificationRequest{},
	)(c)
}

func (h *UserHandler) GetUnreadCount(c *gin.Context) {
	Handle(
		h.Handler,
		func(c *gin.Context, req *EmptyRequest) (*notification.UnreadCountResponse, error) {
			userID := GetUserID(c)
			count, err := h.userService.GetUnreadCount(c.Request.Context(), userID)
			if err != nil {
				return nil, err
			}
			return &notification.UnreadCountResponse{Count: count}, nil
		},
		http.StatusOK,
		&EmptyRequest{},
	)(c)
}
