package notification

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetNotificationRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetNotificationRequest) Validate() error {
	return validate.Struct(r)
}

type ListNotificationsRequest struct {
	model.PaginationRequest
	Unread *bool `form:"unread" filter:"read,eq" binding:"omitempty"`
}

type CreateNotificationRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Message string `json:"message" binding:"required,min=1,max=1000"`
	Type    string `json:"type" binding:"omitempty,oneof=info warning error success"`
}

func (r CreateNotificationRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CreateNotificationRequest) Normalize() {
	r.Title = strings.TrimSpace(r.Title)
	r.Message = strings.TrimSpace(r.Message)
	r.Type = strings.ToLower(strings.TrimSpace(r.Type))
}

type MarkAsReadRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r MarkAsReadRequest) Validate() error {
	return validate.Struct(r)
}

type MarkAllAsReadRequest struct{}

func (r MarkAllAsReadRequest) Validate() error {
	return nil
}

type DeleteNotificationRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r DeleteNotificationRequest) Validate() error {
	return validate.Struct(r)
}

type UnreadCountResponse struct {
	Count int `json:"count"`
}
