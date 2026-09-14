package repository

import (
	"strings"

	"github.com/omar-shahieen/moneyflow/internal/model"
)

func EnsureSortColumn(sort *string, allowed map[string]bool, defaultCol string) string {
	if sort != nil && allowed[*sort] {
		return *sort
	}
	return defaultCol
}

func EnsureSortOrder(order *string) string {
	if order != nil {
		o := strings.ToLower(*order)
		if o == "asc" || o == "desc" {
			return o
		}
	}
	return "desc"
}

func Offset(query *model.ListQuery) int {
	return (query.Page - 1) * query.Limit
}

func NewPaginatedResponse[T any](data []T, query *model.ListQuery, total int) *model.PaginatedResponse[T] {
	totalPages := 0
	if query.Limit > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}
	return &model.PaginatedResponse[T]{
		Data:       data,
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func EmptyPaginatedResponse[T any](query *model.ListQuery) *model.PaginatedResponse[T] {
	return &model.PaginatedResponse[T]{
		Data:       []T{},
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      0,
		TotalPages: 0,
	}
}
