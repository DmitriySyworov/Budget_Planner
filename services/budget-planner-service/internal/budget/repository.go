package budget

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/model"
	"shared/loggers"
	"shared/open_db"
	"time"

	"gorm.io/gorm/clause"
)

type RepositoryBudget struct {
	*open_db.Postgres
	*loggers.Logger
}

func NewRepositoryBudget(db *open_db.Postgres, logger *loggers.Logger) *RepositoryBudget {
	return &RepositoryBudget{
		Postgres: db,
		Logger:   logger,
	}
}

func (r *RepositoryBudget) CreateBudget(budget *model.Budgets) error {
	if errCreate := r.Postgres.Create(&budget).Error; errCreate != nil {
		r.Logger.Error("failed to create budget: " + errCreate.Error())
		return errCreate
	}
	return nil
}
func (r *RepositoryBudget) UpdateBudget(budget *model.Budgets, userUUID, budgetUUID string) error {
	errUpdate := r.Postgres.Clauses(clause.Returning{}).
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Updates(budget).Error
	if errUpdate != nil {
		r.Logger.Error("failed to update budget: " + errUpdate.Error())
		return errUpdate
	}
	return nil
}
func (r *RepositoryBudget) GetBudget(userUUID, budgetUUID string) (*model.Budgets, error) {
	budget := &model.Budgets{}
	if errGet := r.Postgres.
		Where("user_uuid = ? AND  budget_uuid = ?", userUUID, budgetUUID).
		Take(budget).Error; errGet != nil {
		return nil, errGet
	}
	return budget, nil
}
func (r *RepositoryBudget) RemoveBudget(userUUID, budgetUUID string) error {
	errRemove := r.Postgres.
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Budgets{}).Error
	if errRemove != nil {
		r.Logger.Error("failed to remove budget: " + errRemove.Error())
		return errRemove
	}
	return nil
}
func (r *RepositoryBudget) DeleteBudget(userUUID, budgetUUID string) error {
	errDelete := r.Postgres.
		Unscoped().
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Budgets{}).Error
	if errDelete != nil {
		r.Logger.Error("failed to delete budget: " + errDelete.Error())
		return errDelete
	}
	return nil
}
func (r *RepositoryBudget) DateOverlap(start, finish time.Time) bool {
	var isOverlap bool
	if errQuery := r.Raw(`SELECT FROM budgets
				WHERE (start, finish) OVERLAPS (?, ?)`, start, finish).Scan(&isOverlap).Error; errQuery != nil {
		r.Logger.Error("failed to check overlap dates: " + errQuery.Error())
		return true
	}
	return isOverlap
}
func (r *RepositoryBudget) ListBudget(userUUID string, limit, offset int) ([]model.Budgets, error) {
	sliceBudget := make([]model.Budgets, 0, limit)
	if erList := r.Postgres.
		Model(&model.Budgets{}).
		Where("user_uuid = ?", userUUID).
		Order("start").
		Offset(offset).
		Limit(limit).
		Scan(&sliceBudget).Error; erList != nil {
		r.Logger.Error("failed to list budget: " + erList.Error())
		return nil, custom_errors.ErrNotFoundBudget
	}
	if len(sliceBudget) == 0 {
		return nil, custom_errors.ErrNotFoundBudget
	}
	return sliceBudget, nil
}
func (r *RepositoryBudget) BudgetExist(userUUID, budgetUUID string) bool {
	var exist bool
	if errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
	SELECT FROM budgets
	WHERE user_uuid = ? AND budget_uuid = ?)`, userUUID, budgetUUID).Scan(&exist).Error; errQuery != nil {
		r.Logger.Error("failed to check the budget existence: " + errQuery.Error())
		return false
	}
	return exist
}
