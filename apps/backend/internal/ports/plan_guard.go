package ports

import "context"

type PlanGuard interface {
	CheckCategoryLimit(ctx context.Context, userID string) error
	CheckTransactionLimit(ctx context.Context, userID string) error
	CheckBudgetLimit(ctx context.Context, userID string) error
	CheckCSVImportLimit(ctx context.Context, userID string, rowCount int) error
	CheckReportLimit(ctx context.Context, userID string) error
}
