package expense

import (
	"app/budget-planner/internal/model"
	"app/budget-planner/internal/open_db"

	"gorm.io/gorm/clause"
)

type RepositoryExpense struct {
	*open_db.Postgres
}

func NewRepositoryExpense(postgres *open_db.Postgres) *RepositoryExpense {
	return &RepositoryExpense{
		Postgres: postgres,
	}
}
func (r *RepositoryExpense) CreateExpense(expense *model.Expense) error {
	if errCreate := r.Postgres.Create(expense).Error; errCreate != nil {
		return errCreate
	}
	return nil
}
func (r *RepositoryExpense) GetExpense(userUUID, budgetUUID string) (*model.Expense, error) {
	expense := &model.Expense{}
	errGet := r.Postgres.
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).Take(expense).Error
	if errGet != nil {
		return nil, errGet
	}
	return expense, nil
}
func (r *RepositoryExpense) UpdateExpense(expense *model.Expense, userUUID, budgetUUID string) error {
	if errUpdate := r.Postgres.
		Clauses(&clause.Returning{}).
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Updates(expense).Error; errUpdate != nil {
		return errUpdate
	}
	return nil
}
func (r *RepositoryExpense) RemoveExpense(userUUID, budgetUUID string) error {
	if errRemove := r.Postgres.
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Expense{}).Error; errRemove != nil {
		return errRemove
	}
	return nil
}
func (r *RepositoryExpense) DeleteExpense(userUUID, budgetUUID string) error {
	if errDelete := r.Postgres.
		Unscoped().
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Expense{}).Error; errDelete != nil {
		return errDelete
	}
	return nil
}
func (r *RepositoryExpense) ListExpense(userUUID string, limit, offset int) ([]model.Expense, error) {
	sliceExpense := make([]model.Expense, 0, 50)
	errList := r.Postgres.
		Where("user_uuid = ?", userUUID).
		Limit(limit).
		Offset(offset).
		Order("updated_at").
		Scan(sliceExpense).Error
	if errList != nil {
		return sliceExpense, errList
	}
	return nil, errList
}
func (r *RepositoryExpense) ExpenseExist(userUUID, budgetUUID string) bool {
	var exist bool
	errExist := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM expense
				 WHERE user_uuid = ?  AND budget_uuid = ?)`, userUUID, budgetUUID).Error
	if !exist || errExist != nil {
		return false
	}
	return true
}
