package di

type IRepoUser interface {
	IsUserExistsByUUID(userUUID string) bool
}
type IRepoBudget interface {
	BudgetExist(userUUID, budgetUUID string) bool
}
