package user

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type UserAccount struct {
	ID          string    `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Timezone    string    `json:"timezone" db:"timezone"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,min=1,max=100"`
	AvatarURL   string `json:"avatar_url" binding:"omitempty,url,max=2048"`
	Timezone    string `json:"timezone" binding:"omitempty,max=100"`
}

func (r UpdateProfileRequest) Validate() error { return validate.Struct(r) }

func (r *UpdateProfileRequest) Normalize() {
	r.DisplayName = strings.TrimSpace(r.DisplayName)
	r.AvatarURL = strings.TrimSpace(r.AvatarURL)
	r.Timezone = strings.TrimSpace(r.Timezone)
}
