package di

type IRepoBudget interface {
	BudgetExist(userUUID, budgetUUID string) bool
}
type IRepoExpense interface {
	ExpenseExist(budgetUUID, expenseUUID string) bool
}
