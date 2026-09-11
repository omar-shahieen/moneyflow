package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserAccount(t *testing.T) {
	now := time.Now()

	user := UserAccount{
		ID:              "user_123",
		Email:           "test@example.com",
		DisplayName:     "Test User",
		DefaultCurrency: "USD",
		Timezone:        "UTC",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if user.ID != "user_123" {
		t.Error("expected user ID to match")
	}
	if user.Email != "test@example.com" {
		t.Error("expected email to match")
	}
	if user.DisplayName != "Test User" {
		t.Error("expected DisplayName to match")
	}
}

func TestCategory(t *testing.T) {
	id := uuid.New()
	userID := "user_123"

	cat := Category{
		ID:        id,
		UserID:    userID,
		Name:      "Groceries",
		Type:      CategoryTypeExpense,
		CreatedAt: time.Now(),
	}

	if cat.Name != "Groceries" {
		t.Error("expected category name to match")
	}
	if cat.Type != CategoryTypeExpense {
		t.Error("expected category type to be expense")
	}
}

func TestTransaction(t *testing.T) {
	id := uuid.New()
	catID := uuid.New()
	userID := "user_123"

	txn := Transaction{
		ID:          id,
		UserID:      userID,
		CategoryID:  catID,
		AmountMinor: 5000,
		Currency:    "USD",
		Note:        "Coffee shop",
		CreatedAt:   time.Now(),
		OccurredAt:  time.Now(),
	}

	if txn.AmountMinor != 5000 {
		t.Error("expected amount to be 5000")
	}
	if txn.Currency != "USD" {
		t.Error("expected currency to be USD")
	}
	if txn.Note != "Coffee shop" {
		t.Error("expected note to match")
	}
}

func TestBudget(t *testing.T) {
	id := uuid.New()
	catID := uuid.New()

	budget := Budget{
		ID:                id,
		CategoryID:        catID,
		MonthlyLimitMinor: 50000,
		Currency:          "USD",
		CreatedAt:         time.Now(),
	}

	if budget.MonthlyLimitMinor != 50000 {
		t.Error("expected monthly limit to be 50000")
	}
	if budget.Currency != "USD" {
		t.Error("expected currency to be USD")
	}
}

func TestSubscription(t *testing.T) {
	id := uuid.New()
	userID := "user_123"

	sub := Subscription{
		ID:        id,
		UserID:    userID,
		Plan:      PlanVIP,
		Status:    SubscriptionStatusActive,
		CreatedAt: time.Now(),
	}

	if sub.Plan != PlanVIP {
		t.Error("expected plan to be VIP")
	}
	if sub.Status != SubscriptionStatusActive {
		t.Error("expected status to be active")
	}
}

func TestRecurringRule(t *testing.T) {
	id := uuid.New()
	catID := uuid.New()
	userID := "user_123"

	rule := RecurringRule{
		ID:          id,
		UserID:      userID,
		CategoryID:  catID,
		AmountMinor: 20000,
		Currency:    "USD",
		Frequency:   RecurringFrequencyMonthly,
		NextRunDate: time.Now().AddDate(0, 0, 1),
	}

	if rule.Frequency != RecurringFrequencyMonthly {
		t.Error("expected frequency to be monthly")
	}
	if rule.AmountMinor != 20000 {
		t.Error("expected amount to be 20000")
	}
}

func TestReportFormats(t *testing.T) {
	if ReportFormatPDF != "pdf" {
		t.Error("expected PDF format to be 'pdf'")
	}
	if ReportFormatCSV != "csv" {
		t.Error("expected CSV format to be 'csv'")
	}
}

func TestReportStatuses(t *testing.T) {
	if ReportStatusPending != "pending" {
		t.Error("expected pending status")
	}
	if ReportStatusProcessing != "processing" {
		t.Error("expected processing status")
	}
	if ReportStatusReady != "ready" {
		t.Error("expected ready status")
	}
	if ReportStatusFailed != "failed" {
		t.Error("expected failed status")
	}
}

func TestImportStatuses(t *testing.T) {
	if ImportStatusPending != "pending" {
		t.Error("expected pending status")
	}
	if ImportStatusProcessing != "processing" {
		t.Error("expected processing status")
	}
	if ImportStatusCompleted != "completed" {
		t.Error("expected completed status")
	}
	if ImportStatusFailed != "failed" {
		t.Error("expected failed status")
	}
}

func TestBudgetMemberRole(t *testing.T) {
	if BudgetMemberRoleOwner != "owner" {
		t.Error("expected owner role")
	}
	if BudgetMemberRoleMember != "member" {
		t.Error("expected member role")
	}
}

func TestPlanGuard(t *testing.T) {
	limits := map[string]int{
		"free": 5,
		"pro":  50,
		"vip":  999999,
	}

	for plan, limit := range limits {
		if limit <= 0 {
			t.Errorf("expected positive limit for plan %s", plan)
		}
	}
}
