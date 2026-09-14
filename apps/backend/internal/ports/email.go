package ports

type Email interface {
	SendWelcomeEmail(to string, firstName string) error
	SendBudgetAlert(to string, budgetName string, percentageUsed float64) error
	SendMonthlySummary(to string, summaryData map[string]string) error
}
