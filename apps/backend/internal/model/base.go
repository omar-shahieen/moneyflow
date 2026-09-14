package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Sort     string `form:"sort" binding:"omitempty"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
}

func (r *PaginationRequest) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 20
	}
	r.Order = strings.ToLower(r.Order)
	if r.Order != "" && r.Order != "asc" && r.Order != "desc" {
		r.Order = "desc"
	}
}

func (r PaginationRequest) Validate() error {
	r.Normalize()
	return nil
}

func (r PaginationRequest) ToListQuery() *ListQuery {
	r.Normalize()
	var sort, order *string
	if r.Sort != "" {
		sort = &r.Sort
	}
	if r.Order != "" {
		order = &r.Order
	}
	return &ListQuery{
		Page:  r.Page,
		Limit: r.PageSize,
		Sort:  sort,
		Order: order,
	}
}

type ListQuery struct {
	Page   int
	Limit  int
	Search *string
	Sort   *string
	Order  *string
}

type PaginatedResponse[T interface{}] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
