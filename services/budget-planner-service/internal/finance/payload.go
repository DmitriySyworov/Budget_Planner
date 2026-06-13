package finance

type Finance struct {
	*Budget          `json:"budget,omitempty"`
	*Expenses        `json:"expenses,omitempty"`
	*ExpensesPercent `json:"expenses_percent,omitempty"`
}
type Budget struct {
	BudgetInitial               string `json:"budget_initial,omitempty"`
	BudgetBalance               string `json:"budget_balance,omitempty"`
	PredictedAverageSpendPerDay string `json:"predicted_average_spend_per_day,omitempty"`
}
type Expenses struct {
	Health      string `json:"health,omitempty"`
	Sport       string `json:"sport,omitempty"`
	Supermarket string `json:"supermarket,omitempty"`
	Restaurant  string `json:"restaurant,omitempty"`
	Other       string `json:"other,omitempty"`
	Savings     string `json:"savings,omitempty"`
	Investments string `json:"investments,omitempty"`
	Leisure     string `json:"leisure,omitempty"`
}
type ExpensesPercent struct {
	HealthExpensePercent      string `json:"health_expense_percent,omitempty"`
	SportExpensePercent       string `json:"sport_expense_percent,omitempty"`
	SupermarketExpensePercent string `json:"supermarket_expense_percent,omitempty"`
	RestaurantExpensePercent  string `json:"restaurant_expense_percent,omitempty"`
	OtherExpensePercent       string `json:"other_expense_percent,omitempty"`
	SavingsExpensePercent     string `json:"savings_expense_percent,omitempty"`
	InvestmentsExpensePercent string `json:"investments_expense_percent,omitempty"`
	LeisureExpensePercent     string `json:"leisure_expense_percent,omitempty"`
}
