package budget

import (
	"app/budget-planner/internal/model"
	"app/budget-planner/internal/open_db"
	"time"

	"gorm.io/gorm/clause"
)

type RepositoryBudget struct {
	*open_db.Postgres
}

func NewRepositoryBudget(db *open_db.Postgres) *RepositoryBudget {
	return &RepositoryBudget{
		Postgres: db,
	}
}

func (r *RepositoryBudget) CreateBudget(budget *model.Budget) error {
	if errCreate := r.Postgres.Create(&budget).Error; errCreate != nil {
		return errCreate
	}
	return nil
}
func (r *RepositoryBudget) UpdateBudget(budget *model.Budget, userUUID, budgetUUID string) error {
	errUpdate := r.Postgres.Clauses(clause.Returning{}).
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Updates(budget).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}
func (r *RepositoryBudget) GetBudget(userUUID, budgetUUID string) (*model.Budget, error) {
	budget := &model.Budget{}
	errGet := r.Postgres.
		Where("user_uuid = ? AND  budget_uuid = ?", userUUID, budgetUUID).
		Take(budget).Error
	if errGet != nil {
		return nil, errGet
	}
	return budget, nil
}
func (r *RepositoryBudget) RemoveBudget(userUUID, budgetUUID string) error {
	errRemove := r.Postgres.
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Budget{}).Error
	if errRemove != nil {
		return errRemove
	}
	return nil
}
func (r *RepositoryBudget) DeleteBudget(userUUID, budgetUUID string) error {
	errDelete := r.Postgres.
		Unscoped().
		Where("user_uuid = ? AND budget_uuid = ?", userUUID, budgetUUID).
		Delete(&model.Budget{}).Error
	if errDelete != nil {
		return errDelete
	}
	return nil
}
func (r *RepositoryBudget) DateOverlap(start, finish time.Time) bool {
	var isOverlap bool
	errQuery := r.Raw(`SELECT FROM budget
				WHERE (start, finish) OVERLAPS (?, ?)`, start, finish).Scan(&isOverlap).Error
	if !isOverlap || errQuery != nil {
		return false
	}
	return true
}
func (r *RepositoryBudget) ListBudget(userUUID string, limit, offset int) ([]model.Budget, error) {
	sliceBudget := make([]model.Budget, 0, limit)
	erList := r.Postgres.Where("user_uuid = ?", userUUID).
		Order("start").
		Offset(offset).
		Limit(limit).
		Scan(&sliceBudget).Error
	if erList != nil || len(sliceBudget) == 0 {
		return nil, ErrNotFoundBudget
	}
	return sliceBudget, nil
}
func (r *RepositoryBudget) BudgetExist(userUUID, budgetUUID string) bool {
	var exist bool
	errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
	SELECT FROM budget
	WHERE user_uuid = ? AND budget_uuid = ?)`, userUUID, budgetUUID).Scan(&exist).Error
	if exist && errQuery == nil {
		return true
	}
	return false
}
