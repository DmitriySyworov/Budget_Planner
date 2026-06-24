package di

import "app/budget-planner/internal/model"

type IRepoBudget interface {
	BudgetExist(userUUID, budgetUUID string) bool
}
type IRepoExpense interface {
	ExpenseExist(budgetUUID, expenseUUID string) bool
}
type IServiceBudget interface {
	HelperValidateBudget(userUUID, budgetUUID string) (*model.Budget, error)
}
