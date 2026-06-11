package expense

import (
	"app/budget-planner/internal/model"
	"app/budget-planner/internal/open_db"

	"gorm.io/gorm"
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
func (r *RepositoryExpense) UpsertExpense(descriptionExpense *model.DescriptionExpense, budgetUUID, expenseUUID string) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		switch descriptionExpense.Category {
		case "health":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (?, 0.0, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET health = expense.health + excluded.health
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "sport":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, ?, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET sport = expense.sport + excluded.sport
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "supermarket":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, ? , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET supermarket = expense.supermarket + excluded.supermarket
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "restaurant":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET restaurant = expense.restaurant + excluded.restaurant
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "leisure":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET leisure = expense.leisure + excluded.leisure
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "investments":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET investments = expense.investments + excluded.investments
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "savings":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET savings = expense.savings + excluded.savings
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		case "other":
			if errUpsertExpense := tx.Raw(`INSERT INTO expense(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET other = expense.other + excluded.other
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				return errUpsertExpense
			}
		}
		if errCreateDescExp := tx.Create(&descriptionExpense).Error; errCreateDescExp != nil {
			return errCreateDescExp
		}
		return nil
	})
}
func (r *RepositoryExpense) GetDescriptionExpense(expenseUUID, descriptionExpenseUUID string) (*model.DescriptionExpense, error) {
	expense := &model.DescriptionExpense{}
	errGet := r.Postgres.
		Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).Take(expense).Error
	if errGet != nil {
		return nil, errGet
	}
	return expense, nil
}
func (r *RepositoryExpense) GetExpenseUUID(budgetUUID string) (string, error) {
	var expenseUUID string
	errGet := r.Postgres.Raw(`SELECT expense_uuid FROM expense
                                   WHERE budget_uuid = ?`, budgetUUID).Scan(&expenseUUID).Error
	if errGet != nil {
		return "", errGet
	}
	return expenseUUID, nil
}

func (r *RepositoryExpense) UpdateDescriptionExpense(expense *model.DescriptionExpense, expenseUUID, descriptionExpenseUUID string) error {
	if errUpdate := r.Postgres.
		Clauses(&clause.Returning{}).
		Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).
		Updates(expense).Error; errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (r *RepositoryExpense) UpdateExpenseTransaction(descriptionExpense *model.DescriptionExpense, oldExpense, budgetUUID, expenseUUID string) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		if errUpdateDescription := r.Postgres.
			Clauses(&clause.Returning{}).
			Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpense.DescriptionExpenseUUID).
			Updates(descriptionExpense).Error; errUpdateDescription != nil {
			return errUpdateDescription
		}
		if errUpdate := r.Postgres.Model(&model.Expense{}).
			Where("expense_uuid = ? AND budget_uuid = ?", expenseUUID, budgetUUID).
			Update(descriptionExpense.Category,
				gorm.Expr("? - ? + ?", descriptionExpense.Category, oldExpense, descriptionExpense.Expense)).
			Error; errUpdate != nil {
			return errUpdate
		}
		return nil
	})
}

type deleteExpenseParams struct {
	categoryExpense        string
	expense                string
	budgetUUID             string
	expenseUUID            string
	descriptionExpenseUUID string
}

func (r *RepositoryExpense) DeleteExpense(params *deleteExpenseParams) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		if errDelete := r.Postgres.Where("expense_uuid = ? AND description_expense_uuid = ?", params.expenseUUID, params.descriptionExpenseUUID).
			Delete(&model.DescriptionExpense{}).Error; errDelete != nil {
			return errDelete
		}
		if errUpdate := r.Postgres.Model(&model.Expense{}).
			Where("expense_uuid = ? AND budget_uuid = ?", params.expenseUUID, params.budgetUUID).
			Update(params.categoryExpense, gorm.Expr("? - ?", params.categoryExpense, params.expense)).
			Error; errUpdate != nil {
			return errUpdate
		}
		return nil
	})
}
func (r *RepositoryExpense) ListExpense(expenseUUID string, limit, offset int) ([]model.DescriptionExpense, error) {
	sliceDescriptionExpense := make([]model.DescriptionExpense, 0, 50)
	errList := r.Postgres.
		Model(&model.DescriptionExpense{}).
		Where("expense_uuid = ?", expenseUUID).
		Limit(limit).
		Offset(offset).
		Order("created_at").
		Scan(sliceDescriptionExpense).Error
	if errList != nil {
		return sliceDescriptionExpense, errList
	}
	return nil, errList
}
func (r *RepositoryExpense) ExpenseExist(budgetUUID, expenseUUID string) bool {
	var exist bool
	if errExist := r.Postgres.Raw(`SELECT EXISTS(
SELECT FROM expense
WHERE budget_uuid = ? AND expense_uuid = ?)`, budgetUUID, expenseUUID).Scan(&exist).
		Error; errExist != nil || !exist {
		return false
	}
	return true
}
