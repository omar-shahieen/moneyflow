package user

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type GetUserRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (r GetUserRequest) Validate() error {
	return validate.Struct(r)
}

type CreateUserRequest struct {
	Email           string `json:"email" binding:"required,email"`
	DisplayName     string `json:"display_name" binding:"required,min=1,max=100"`
	DefaultCurrency string `json:"default_currency" binding:"required,len=3"`
	Timezone        string `json:"timezone" binding:"required"`
}

func (r CreateUserRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateUserRequest struct {
	DisplayName     string `json:"display_name" binding:"required,min=1,max=100"`
	AvatarURL       string `json:"avatar_url" binding:"omitempty,url"`
	DefaultCurrency string `json:"default_currency" binding:"required,len=3"`
	Timezone        string `json:"timezone" binding:"required"`
}

func (r UpdateUserRequest) Validate() error {
	return validate.Struct(r)
}

type UserResponse struct {
	UserAccount
}

type UserListResponse struct {
	Data       []UserResponse `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int            `json:"total"`
	TotalPages int            `json:"totalPages"`
}
