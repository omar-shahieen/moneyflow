package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		Data: data, Page: query.Page, Limit: query.Limit, Total: total, TotalPages: totalPages,
	}
}

func EmptyPaginatedResponse[T any](query *model.ListQuery) *model.PaginatedResponse[T] {
	return &model.PaginatedResponse[T]{
		Data: []T{}, Page: query.Page, Limit: query.Limit, Total: 0, TotalPages: 0,
	}
}

// FilterApplier lets a request struct contribute conditions that aren't a
// simple "column op value" (subqueries, joins).
type FilterApplier interface {
	ApplyCustomFilters(args pgx.NamedArgs) string
}

// BuildFilterClause reflects over src (a request struct — PaginationRequest's
// embedded fields are skipped, they carry no `filter` tag) and appends a
// WHERE condition for every non-nil pointer field tagged `filter:"column,op"`.
// Supported ops: eq, ilike, gte, lte.
func BuildFilterClause(src any, args pgx.NamedArgs) string {
	if src == nil {
		return ""
	}
	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}

	var clauses []string
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if sub := BuildFilterClause(v.Field(i).Addr().Interface(), args); sub != "" {
				clauses = append(clauses, strings.TrimPrefix(sub, " AND "))
			}
			continue
		}

		tag := field.Tag.Get("filter")
		if tag == "" {
			continue
		}
		parts := strings.SplitN(tag, ",", 2)
		column, op := parts[0], "eq"
		if len(parts) == 2 {
			op = parts[1]
		}

		fv := v.Field(i)
		if fv.Kind() != reflect.Ptr || fv.IsNil() {
			continue
		}

		argName := column + "_" + strconv.Itoa(i)
		value := fv.Elem().Interface()

		switch op {
		case "eq":
			clauses = append(clauses, fmt.Sprintf("%s = @%s", column, argName))
		case "ilike":
			clauses = append(clauses, fmt.Sprintf("%s ILIKE '%%' || @%s || '%%'", column, argName))
		case "gte":
			clauses = append(clauses, fmt.Sprintf("%s >= @%s", column, argName))
		case "lte":
			clauses = append(clauses, fmt.Sprintf("%s <= @%s", column, argName))
		default:
			continue
		}
		args[argName] = value
	}

	if applier, ok := src.(FilterApplier); ok {
		if extra := applier.ApplyCustomFilters(args); extra != "" {
			clauses = append(clauses, strings.TrimPrefix(extra, " AND "))
		}
	}

	if len(clauses) == 0 {
		return ""
	}
	return " AND " + strings.Join(clauses, " AND ")
}

// RunPaginatedQuery executes count + select + collect + paginate in one place.
// baseStmt/countStmt must already include the full WHERE clause but not
// ORDER BY / LIMIT / OFFSET.
func RunPaginatedQuery[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	baseStmt, countStmt string,
	args pgx.NamedArgs,
	query *model.ListQuery,
	allowedSort map[string]bool,
	defaultSort string,
) (*model.PaginatedResponse[T], error) {
	var total int
	if err := pool.QueryRow(ctx, countStmt, args).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}
	if total == 0 {
		return EmptyPaginatedResponse[T](query), nil
	}

	sortColumn := EnsureSortColumn(query.Sort, allowedSort, defaultSort)
	sortOrder := EnsureSortOrder(query.Order)

	stmt := baseStmt + fmt.Sprintf(" ORDER BY %s %s LIMIT @limit OFFSET @offset", sortColumn, sortOrder)
	args["limit"] = query.Limit
	args["offset"] = Offset(query)

	rows, err := pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmptyPaginatedResponse[T](query), nil
		}
		return nil, fmt.Errorf("failed to collect rows: %w", err)
	}

	return NewPaginatedResponse(data, query, total), nil
}
