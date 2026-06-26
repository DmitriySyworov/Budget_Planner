package finance

import (
	"shared/loggers"
	"shared/open_db"
)

type RepositoryFinance struct {
	*open_db.Postgres
	*loggers.Logger
}

func NewRepositoryFinance(postgres *open_db.Postgres, logger *loggers.Logger) *RepositoryFinance {
	return &RepositoryFinance{
		Postgres: postgres,
	}
}

type DTOFinance struct {
	BudgetInitial               string `gorm:"column:budget_initial"`
	BudgetBalance               string `gorm:"column:budget_balance"`
	PredictedAverageSpendPerDay string `gorm:"column:predicted_average_spend_per_day"`
	Health                      string `gorm:"column:health"`
	Sport                       string `gorm:"column:sport"`
	Supermarket                 string `gorm:"column:supermarket"`
	Restaurant                  string `gorm:"column:restaurant"`
	Other                       string `gorm:"column:other"`
	Savings                     string `gorm:"column:savings"`
	Investments                 string `gorm:"column:investments"`
	Leisure                     string `gorm:"column:leisure"`
	HealthExpensePercent        string `gorm:"column:health_expense_percent"`
	SportExpensePercent         string `gorm:"column:sport_expense_percent"`
	SupermarketExpensePercent   string `gorm:"column:supermarket_expense_percent"`
	RestaurantExpensePercent    string `gorm:"column:restaurant_expense_percent"`
	OtherExpensePercent         string `gorm:"column:other_expense_percent"`
	SavingsExpensePercent       string `gorm:"column:savings_expense_percent"`
	InvestmentsExpensePercent   string `gorm:"column:investments_expense_percent"`
	LeisureExpensePercent       string `gorm:"column:leisure_expense_percent"`
}

func (r *RepositoryFinance) Finance(budgetUUID, expenseUUID string) (*DTOFinance, error) {
	dtoFinance := &DTOFinance{}
	if errQueryFinance := r.Postgres.Raw(`WITH calculation_expense as (
SELECT  expenses.other + expenses.supermarket +  expenses.restaurant + expenses.health + expenses.sport + expenses.savings + expenses.investments + expenses.leisure  as sum_expsense,
    expenses.*  FROM expenses
    WHERE budget_uuid = ? AND expense_uuid = ?
),
calculation_descrition_expense as (
SELECT
COUNT(*) FILTER ( WHERE category = 'health') as counter_expense_health,
COUNT(*) FILTER ( WHERE category = 'other') as counter_expense_other,
COUNT(*) FILTER ( WHERE category = 'supermarket') as counter_expense_supermarket,
COUNT(*) FILTER ( WHERE category = 'restaurant' ) as counter_expense_restaurant,
COUNT(*) FILTER ( WHERE category = 'savings') as counter_expense_savings,
COUNT(*) FILTER ( WHERE category = 'sport') as counter_expense_sport,
COUNT(*) FILTER ( WHERE category = 'investments') as counter_expense_investments,
COUNT(*) FILTER ( WHERE category = 'leisure' ) as counter_expense_leisure,
expense_uuid
FROM description_expenses
WHERE expense_uuid = ?
GROUP BY  expense_uuid
)
SELECT
    e.health, e.other, e.supermarket, e.restaurant, e.sport, e.leisure, e.savings, e.investments,
    d.counter_expense_health,
    d.counter_expense_other,
    d.counter_expense_supermarket,
    d.counter_expense_restaurant,
    d.counter_expense_sport,
    d.counter_expense_leisure,
    d.counter_expense_savings,
    d.counter_expense_investments,
    budgets.amount - e.sum_expsense                                        as budget_balance,
    budgets.amount as initial_budget,
    ROUND(e.health * 100.0 / NULLIF(e.sum_expsense, 0), 2)      as health_expense_percent,
    ROUND(e.other * 100.0 / NULLIF(e.sum_expsense, 0), 2)       as other_expense_percent,
    ROUND(e.supermarket * 100.0 / NULLIF(e.sum_expsense, 0), 2) as supermarket_expense_percent,
    ROUND(e.restaurant * 100.0 / NULLIF(e.sum_expsense, 0), 2)  as restaurant_expense_percent,
    ROUND(e.sport * 100.0 / NULLIF(e.sum_expsense, 0), 2)       as sport_expense_percent,
    ROUND(e.leisure * 100.0 / NULLIF(e.sum_expsense, 0), 2)     as leisure_expense_percent,
    ROUND(e.savings * 100.0 / NULLIF(e.sum_expsense, 0), 2)     as savings_expense_percent,
    ROUND(e.investments * 100.0 / NULLIF(e.sum_expsense, 0), 2) as investments_expense_percent,
    ROUND(budgets.amount / NULLIF(budgets.finish-budgets.start, 0), 2)  as predicted_average_spend_per_day
FROM calculation_expense e
JOIN budgets ON budgets.budget_uuid = e.budget_uuid
LEFT JOIN  calculation_descrition_expense d ON d.expense_uuid = e.expense_uuid
`, budgetUUID, expenseUUID, expenseUUID).Scan(dtoFinance).Error; errQueryFinance != nil {
		r.Logger.Error("failed to get user finance: " + errQueryFinance.Error())
		return nil, errQueryFinance
	}
	return dtoFinance, nil
}
