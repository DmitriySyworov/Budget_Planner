package expense

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/model"
	"shared/loggers"
	"shared/open_db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepositoryExpense struct {
	*open_db.Postgres
	*loggers.Logger
}

func NewRepositoryExpense(postgres *open_db.Postgres, logger *loggers.Logger) *RepositoryExpense {
	return &RepositoryExpense{
		Postgres: postgres,
		Logger:   logger,
	}
}
func (r *RepositoryExpense) UpsertExpense(descriptionExpense *model.DescriptionExpenses, budgetUUID, expenseUUID string) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		switch descriptionExpense.Category {
		case "health":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (?, 0.0, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET health = expenses.health + excluded.health
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "sport":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, ?, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET sport = expenses.sport + excluded.sport
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "supermarket":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, ? , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET supermarket = expenses.supermarket + excluded.supermarket
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "restaurant":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET restaurant = expenses.restaurant + excluded.restaurant
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "leisure":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET leisure = expenses.leisure + excluded.leisure
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "investments":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET investments = expenses.investments + excluded.investments
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "savings":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET savings = expenses.savings + excluded.savings
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "other":
			if errUpsertExpense := tx.Raw(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET other = expenses.other + excluded.other
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		}
		if errCreateDescExp := tx.Create(&descriptionExpense).Error; errCreateDescExp != nil {
			r.Logger.Error("failed to create description_expense: " + errCreateDescExp.Error())
			return errCreateDescExp
		}
		return nil
	})
}
func (r *RepositoryExpense) GetDescriptionExpense(expenseUUID, descriptionExpenseUUID string) (*model.DescriptionExpenses, error) {
	expense := &model.DescriptionExpenses{}
	errGet := r.Postgres.
		Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).Take(expense).Error
	if errGet != nil {
		return nil, errGet
	}
	return expense, nil
}
func (r *RepositoryExpense) GetExpenseUUID(budgetUUID string) (string, error) {
	var expenseUUID string
	if errGet := r.Postgres.Raw(`SELECT expense_uuid FROM expenses
                                   WHERE budget_uuid = ?`, budgetUUID).Scan(&expenseUUID).Error; errGet != nil {
		r.Logger.Error("failed to get expense_uuid: " + errGet.Error())
		return "", errGet
	}
	if expenseUUID == "" {
		return "", custom_errors.ErrNotFoundExpense
	}
	return expenseUUID, nil
}

func (r *RepositoryExpense) UpdateDescriptionExpense(expense *model.DescriptionExpenses, expenseUUID, descriptionExpenseUUID string) error {
	if errUpdate := r.Postgres.
		Clauses(&clause.Returning{}).
		Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).
		Updates(expense).Error; errUpdate != nil {
		r.Logger.Error("failed to update description expense: " + errUpdate.Error())
		return errUpdate
	}
	return nil
}

func (r *RepositoryExpense) UpdateExpenseTransaction(descriptionExpense *model.DescriptionExpenses, oldExpense, budgetUUID, expenseUUID string) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		if errUpdateDescription := r.Postgres.
			Clauses(&clause.Returning{}).
			Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpense.DescriptionExpenseUUID).
			Updates(descriptionExpense).Error; errUpdateDescription != nil {
			r.Logger.Error("failed to update description: " + errUpdateDescription.Error())
			return errUpdateDescription
		}
		if errUpdate := r.Postgres.Model(&model.Expenses{}).
			Where("expense_uuid = ? AND budget_uuid = ?", expenseUUID, budgetUUID).
			Update(descriptionExpense.Category,
				gorm.Expr("? - ? + ?", descriptionExpense.Category, oldExpense, descriptionExpense.Expense)).
			Error; errUpdate != nil {
			r.Logger.Error("failed to add expense: " + errUpdate.Error())
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
			Delete(&model.DescriptionExpenses{}).Error; errDelete != nil {
			r.Logger.Error("failed to delete expense: " + errDelete.Error())
			return errDelete
		}
		if errUpdate := r.Postgres.Model(&model.Expenses{}).
			Where("expense_uuid = ? AND budget_uuid = ?", params.expenseUUID, params.budgetUUID).
			Update(params.categoryExpense, gorm.Expr("? - ?", params.categoryExpense, params.expense)).
			Error; errUpdate != nil {
			r.Logger.Error("failed to update expense: " + errUpdate.Error())
			return errUpdate
		}
		return nil
	})
}
func (r *RepositoryExpense) ListExpense(expenseUUID string, limit, offset int) ([]model.DescriptionExpenses, error) {
	sliceDescriptionExpense := make([]model.DescriptionExpenses, 0, 50)
	if errList := r.Postgres.
		Model(&model.DescriptionExpenses{}).
		Where("expense_uuid = ?", expenseUUID).
		Limit(limit).
		Offset(offset).
		Order("created_at").
		Scan(&sliceDescriptionExpense).Error; errList != nil {
		r.Logger.Error("failed to list description_expense: " + errList.Error())
		return nil, errList
	}
	if len(sliceDescriptionExpense) == 0 {
		return nil, ErrNotFoundDescriptionExpense
	}
	return sliceDescriptionExpense, nil
}
func (r *RepositoryExpense) ExpenseExist(budgetUUID, expenseUUID string) bool {
	var exist bool
	if errExist := r.Postgres.Raw(`SELECT EXISTS(
SELECT FROM expenses
WHERE budget_uuid = ? AND expense_uuid = ?)`, budgetUUID, expenseUUID).Scan(&exist).
		Error; errExist != nil {
		r.Logger.Error("failed to check the expense existence: " + errExist.Error())
		return false
	}
	return exist
}
