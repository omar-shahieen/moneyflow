package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/ports"
)

type NotificationService struct {
	pool  *pgxpool.Pool
	email ports.Email
}

func NewNotificationService(pool *pgxpool.Pool, email ports.Email) *NotificationService {
	return &NotificationService{
		pool:  pool,
		email: email,
	}
}

type BudgetAlertData struct {
	BudgetName     string
	CategoryName   string
	PercentageUsed float64
	MonthlyLimit   int64
	CurrentUsage   int64
}

func (s *NotificationService) CheckBudgetAlerts(ctx context.Context) error {
	rows, err := s.pool.Query(ctx,
		`SELECT u.id, u.email, u.display_name,
		        b.monthly_limit_minor, c.name as category_name,
		        COALESCE(SUM(t.amount_minor), 0) as current_usage
		 FROM budgets b
		 JOIN budget_members bm ON b.id = bm.budget_id
		 JOIN user_accounts u ON bm.user_id = u.id
		 JOIN categories c ON b.category_id = c.id
		 LEFT JOIN transactions t ON t.category_id = b.category_id
		   AND t.user_id = bm.user_id
		   AND t.occurred_at >= date_trunc('month', NOW())
		   AND t.occurred_at < date_trunc('month', NOW()) + INTERVAL '1 month'
		 WHERE bm.role = 'owner'
		 GROUP BY u.id, u.email, u.display_name, b.id, b.monthly_limit_minor, c.name
		 HAVING COALESCE(SUM(t.amount_minor), 0) > b.monthly_limit_minor * 0.8`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, email, displayName, categoryName string
		var monthlyLimit, currentUsage int64

		if err := rows.Scan(&userID, &email, &displayName, &monthlyLimit, &categoryName, &currentUsage); err != nil {
			continue
		}

		percentageUsed := float64(currentUsage) / float64(monthlyLimit) * 100

		if s.email != nil {
			_ = s.email.SendBudgetAlert(email, categoryName, percentageUsed)
		}
	}

	return nil
}

func (s *NotificationService) SendMonthlySummaries(ctx context.Context) error {
	rows, err := s.pool.Query(ctx,
		`SELECT u.id, u.email, u.display_name
		 FROM user_accounts u
		 WHERE u.id IN (
		     SELECT DISTINCT user_id FROM transactions
		     WHERE occurred_at >= date_trunc('month', NOW() - INTERVAL '1 month')
		       AND occurred_at < date_trunc('month', NOW())
		 )`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, email, displayName string
		if err := rows.Scan(&userID, &email, &displayName); err != nil {
			continue
		}

		summary, err := s.getMonthlySummary(ctx, userID)
		if err != nil {
			continue
		}

		if s.email != nil {
			_ = s.email.SendMonthlySummary(email, summary)
		}
	}

	return nil
}

func (s *NotificationService) getMonthlySummary(ctx context.Context, userID string) (map[string]string, error) {
	lastMonth := time.Now().AddDate(0, -1, 0)
	startOfMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var totalIncome, totalExpense int64
	err := s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount_minor ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount_minor ELSE 0 END), 0)
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1 AND t.occurred_at >= $2 AND t.occurred_at < $3`,
		userID, startOfMonth, endOfMonth,
	).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Month":        lastMonth.Format("January 2006"),
		"TotalIncome":  fmt.Sprintf("%d", totalIncome),
		"TotalExpense": fmt.Sprintf("%d", totalExpense),
		"NetSavings":   fmt.Sprintf("%d", totalIncome-totalExpense),
	}, nil
}
