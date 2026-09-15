package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/errs"
	"github.com/omar-shahieen/moneyflow/internal/model"
	"github.com/omar-shahieen/moneyflow/internal/model/notification"
	"github.com/omar-shahieen/moneyflow/internal/model/user"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
)

type UserService struct {
	server           *server.Server
	userRepo         *repository.UserRepo
	notificationRepo *repository.NotificationRepo
}

func NewUserService(server *server.Server, userRepo *repository.UserRepo, notificationRepo *repository.NotificationRepo) *UserService {
	return &UserService{
		server:           server,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*user.UserAccount, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to fetch user profile")
		return nil, errs.NewNotFoundError("user not found", false, nil)
	}
	return u, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, payload *user.UpdateProfileRequest) (*user.UserAccount, error) {
	existing, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to fetch user for update")
		return nil, errs.NewNotFoundError("user not found", false, nil)
	}

	if payload.DisplayName != "" {
		existing.DisplayName = payload.DisplayName
	}
	if payload.AvatarURL != "" {
		existing.AvatarURL = payload.AvatarURL
	}
	if payload.Timezone != "" {
		existing.Timezone = payload.Timezone
	}

	if err := s.userRepo.Update(ctx, existing); err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to update user profile")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "user_profile_updated").
		Str("user_id", userID).
		Msg("User profile updated successfully")

	return existing, nil
}

func (s *UserService) GetNotifications(ctx context.Context, userID string, req *notification.ListNotificationsRequest) (*model.PaginatedResponse[notification.Notification], error) {
	notifications, err := s.notificationRepo.List(ctx, userID, req)
	if err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to fetch notifications")
		return nil, err
	}
	return notifications, nil
}

func (s *UserService) GetNotificationByID(ctx context.Context, userID string, notificationID uuid.UUID) (*notification.Notification, error) {
	n, err := s.notificationRepo.GetByID(ctx, notificationID, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to fetch notification")
		return nil, errs.NewNotFoundError("notification not found", false, nil)
	}
	return n, nil
}

func (s *UserService) CreateNotification(ctx context.Context, userID string, payload *notification.CreateNotificationRequest) (*notification.Notification, error) {
	n := &notification.Notification{
		UserID:  userID,
		Title:   payload.Title,
		Message: payload.Message,
		Type:    payload.Type,
	}

	if n.Type == "" {
		n.Type = "info"
	}

	if err := s.notificationRepo.Create(ctx, n); err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to create notification")
		return nil, err
	}

	s.server.Logger.Info().
		Str("event", "notification_created").
		Str("user_id", userID).
		Str("notification_id", n.ID.String()).
		Msg("Notification created successfully")

	return n, nil
}

func (s *UserService) MarkNotificationAsRead(ctx context.Context, userID string, notificationID uuid.UUID) error {
	if err := s.notificationRepo.MarkAsRead(ctx, notificationID, userID); err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to mark notification as read")
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (s *UserService) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to mark all notifications as read")
		return err
	}
	return nil
}

func (s *UserService) DeleteNotification(ctx context.Context, userID string, notificationID uuid.UUID) error {
	if err := s.notificationRepo.Delete(ctx, notificationID, userID); err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to delete notification")
		return err
	}
	return nil
}

func (s *UserService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count, err := s.notificationRepo.UnreadCount(ctx, userID)
	if err != nil {
		s.server.Logger.Error().Err(err).Str("user_id", userID).Msg("failed to get unread notification count")
		return 0, err
	}
	return count, nil
}
