package category

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/omar-shahieen/moneyflow/internal/model"
)

var validate = validator.New()

type GetCategoryRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r GetCategoryRequest) Validate() error {
	return validate.Struct(r)
}

type ListCategoriesRequest struct {
	model.PaginationRequest
	Type string `form:"type" binding:"omitempty,oneof=income expense"`
}

type CategoryFilters struct {
	Type *string
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	Type string `json:"type" binding:"required,oneof=income expense"`
}

func (r CreateCategoryRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateCategoryRequest struct {
	ID   uuid.UUID `uri:"id" binding:"required,uuid"`
	Name string    `json:"name" binding:"required,min=1,max=100"`
	Type string    `json:"type" binding:"required,oneof=income expense"`
}

func (r UpdateCategoryRequest) Validate() error {
	return validate.Struct(r)
}

type DeleteCategoryRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

func (r DeleteCategoryRequest) Validate() error {
	return validate.Struct(r)
}

type CategoryResponse struct {
	Category
}
