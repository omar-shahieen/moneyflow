package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

type BaseWithId struct {
	ID uuid.UUID `json:"id" db:"id"`
}

type BaseWithCreatedAt struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type BaseWithUpdatedAt struct {
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Base struct {
	BaseWithId
	BaseWithCreatedAt
	BaseWithUpdatedAt
}

type PaginationRequest struct {
	Page     *int    `form:"page" binding:"omitempty,min=1"`
	PageSize *int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Sort     *string `form:"sort" binding:"omitempty"`
	Order    *string `form:"order" binding:"omitempty,oneof=asc desc"`
	Search   *string `form:"search" binding:"omitempty,max=200"`
}

func (r *PaginationRequest) Normalize() {
	if r.Page == nil {
		page := 1
		r.Page = &page
	}

	if r.PageSize == nil {
		pageSize := 20
		r.PageSize = &pageSize
	}

	if r.Order == nil {
		order := "dsc"
		r.Order = &order
	}
	if r.Sort == nil {
		Sort := "created_at"
		r.Sort = &Sort
	}
}

func (r PaginationRequest) Validate() error {
	r.Normalize()
	return validate.Struct(r)
}

type ListQuery struct {
	Page   int
	Limit  int
	Search *string
	Sort   *string
	Order  *string
}

func (r PaginationRequest) ToListQuery() *ListQuery {
	r.Normalize()
	return &ListQuery{
		Page:   *r.Page,
		Limit:  *r.PageSize,
		Search: r.Search,
		Sort:   r.Sort,
		Order:  r.Order,
	}
}

type PaginatedResponse[T interface{}] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
