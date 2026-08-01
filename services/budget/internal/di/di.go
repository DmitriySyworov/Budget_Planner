package di

import (
	"app/budget-planner/internal/model"
	"context"
)

type IRepoBudget interface {
	BudgetExist(ctxRequest context.Context, userUUID, budgetUUID string) bool
}
type IRepoExpense interface {
	ExpenseExist(ctxRequest context.Context, budgetUUID, expenseUUID string) bool
}
type IServiceBudget interface {
	HelperValidateBudget(ctxRequest context.Context, userUUID, budgetUUID string) (*model.Budgets, error)
}
