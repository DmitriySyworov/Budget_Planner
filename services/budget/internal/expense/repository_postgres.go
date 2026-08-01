package expense

import (
	"app/budget-planner/internal/apperrors"
	"app/budget-planner/internal/model"
	"context"
	"shared/loggers"
	"shared/storage"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IRepositoryExpense interface {
	UpsertExpense(ctxRequest context.Context, descriptionExpense *model.DescriptionExpenses, budgetUUID, expenseUUID string) error
	GetDescriptionExpense(ctxRequest context.Context, expenseUUID, descriptionExpenseUUID string) (*model.DescriptionExpenses, error)
	GetExpense(ctxRequest context.Context, budgetUUID string) (*model.Expenses, error)
	GetExpenseUUID(ctxRequest context.Context, budgetUUID string) (string, error)
	UpdateDescriptionExpense(ctxRequest context.Context, expense *model.DescriptionExpenses) error
	UpdateExpenseTransaction(ctxRequest context.Context, descriptionExpense *model.DescriptionExpenses, oldExpense, oldCategory, budgetUUID, expenseUUID string) error
	DeleteDescriptionExpense(ctxRequest context.Context, params *deleteExpenseParams) error
	ListDescriptionExpense(ctxRequest context.Context, expenseUUID string, limit, offset int) ([]model.DescriptionExpenses, error)
	ExpenseExist(ctxRequest context.Context, budgetUUID, expenseUUID string) bool
}

type RepositoryExpense struct {
	*storage.Postgres
	*loggers.Logger
}

func NewRepositoryExpense(postgres *storage.Postgres, logger *loggers.Logger) IRepositoryExpense {
	return &RepositoryExpense{
		Postgres: postgres,
		Logger:   logger,
	}
}
func (r *RepositoryExpense) UpsertExpense(ctxRequest context.Context, descriptionExpense *model.DescriptionExpenses, budgetUUID, expenseUUID string) error {
	return r.Postgres.Transaction(func(tx *gorm.DB) error {
		switch descriptionExpense.Category {
		case "health":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (?, 0.0, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET health = expenses.health + excluded.health
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "sport":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, ?, 0.0 , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET sport = expenses.sport + excluded.sport
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "supermarket":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, ? , 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET supermarket = expenses.supermarket + excluded.supermarket
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "restaurant":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET restaurant = expenses.restaurant + excluded.restaurant
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "leisure":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET leisure = expenses.leisure + excluded.leisure
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "investments":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET investments = expenses.investments + excluded.investments
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "savings":
			if errUpsertExpense := tx.Exec(`INSERT INTO expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, 0.0, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET savings = expenses.savings + excluded.savings
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		case "other":
			if errUpsertExpense := tx.Exec(`INSERT INTO  expenses(health, sport, supermarket, restaurant, leisure, investments, savings, other, budget_uuid, expense_uuid) 
				VALUES (0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, ?, ?, ?)
				ON CONFLICT (expense_uuid)
				DO UPDATE SET other = expenses.other + excluded.other
				`, descriptionExpense.Expense, budgetUUID, expenseUUID).Error; errUpsertExpense != nil {
				r.Logger.Error("failed to upsert expense: " + errUpsertExpense.Error())
				return errUpsertExpense
			}
		}
		if errCreateDescExp := tx.WithContext(ctxRequest).Create(&descriptionExpense).Error; errCreateDescExp != nil {
			r.Logger.Error("failed to create description_expense: " + errCreateDescExp.Error())
			return errCreateDescExp
		}
		return nil
	})
}
func (r *RepositoryExpense) GetDescriptionExpense(ctxRequest context.Context, expenseUUID, descriptionExpenseUUID string) (*model.DescriptionExpenses, error) {
	descriptionExpense := &model.DescriptionExpenses{}
	errGet := r.Postgres.
		WithContext(ctxRequest).
		Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).
		Take(descriptionExpense).Error
	if errGet != nil {
		return nil, errGet
	}
	return descriptionExpense, nil
}
func (r *RepositoryExpense) GetExpense(ctxRequest context.Context, budgetUUID string) (*model.Expenses, error) {
	expense := &model.Expenses{}
	errGet := r.Postgres.
		WithContext(ctxRequest).
		Where("budget_uuid = ?", budgetUUID).Take(expense).Error
	if errGet != nil {
		return nil, errGet
	}
	return expense, nil
}
func (r *RepositoryExpense) GetExpenseUUID(ctxRequest context.Context, budgetUUID string) (string, error) {
	var expenseUUID string
	if errGet := r.Postgres.
		WithContext(ctxRequest).
		Raw(`SELECT expense_uuid FROM expenses
				  WHERE budget_uuid = ?`, budgetUUID).Scan(&expenseUUID).Error; errGet != nil {
		r.Logger.Error("failed to get expense_uuid: " + errGet.Error())
		return "", errGet
	}
	if expenseUUID == "" {
		return "", apperrors.ErrNotFoundExpense
	}
	return expenseUUID, nil
}

func (r *RepositoryExpense) UpdateDescriptionExpense(ctxRequest context.Context, expense *model.DescriptionExpenses) error {
	if errUpdate := r.Postgres.
		WithContext(ctxRequest).
		Clauses(&clause.Returning{}).
		Where("expense_uuid = ? AND description_expense_uuid = ?", expense.ExpenseUUID, expense.DescriptionExpenseUUID).
		Updates(expense).Error; errUpdate != nil {
		r.Logger.Error("failed to update description expense: " + errUpdate.Error())
		return errUpdate
	}
	return nil
}

func (r *RepositoryExpense) UpdateExpenseTransaction(ctxRequest context.Context, descriptionExpense *model.DescriptionExpenses, oldExpense, oldCategory, budgetUUID, expenseUUID string) error {
	return r.Postgres.WithContext(ctxRequest).Transaction(func(tx *gorm.DB) error {
		if errUpdateDescription := tx.
			Clauses(&clause.Returning{}).
			Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpense.DescriptionExpenseUUID).
			Updates(descriptionExpense).Error; errUpdateDescription != nil {
			r.Logger.Error("failed to update description_expense: " + errUpdateDescription.Error())
			return errUpdateDescription
		}
		var queryUpdateExpense string
		if descriptionExpense.Category != "" && descriptionExpense.Category != oldCategory {
			queryUpdateExpense = "UPDATE expenses" +
				"SET " + oldCategory + " = " + oldCategory + " :: numeric - " + oldExpense + " ::numeric," +
				descriptionExpense.Category + " = " + descriptionExpense.Category + " ::numeric + " + descriptionExpense.Expense + " ::numeric"

		} else {
			queryUpdateExpense = "UPDATE expenses SET " +
				oldCategory + " = " + oldCategory + " ::numeric - " + oldExpense + " ::numeric + " + descriptionExpense.Expense + " ::numeric"
		}
		queryUpdateExpense += " WHERE expense_uuid = '" + descriptionExpense.ExpenseUUID + "' AND budget_uuid = '" + budgetUUID + "'"
		if errUpdateExpense := tx.Exec(queryUpdateExpense).Error; errUpdateExpense != nil {
			r.Logger.Error("failed to update expense: " + errUpdateExpense.Error())
			return errUpdateExpense
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

func (r *RepositoryExpense) DeleteDescriptionExpense(ctxRequest context.Context, params *deleteExpenseParams) error {
	return r.Postgres.WithContext(ctxRequest).Transaction(func(tx *gorm.DB) error {
		if errDelete := tx.Where("expense_uuid = ? AND description_expense_uuid = ?", params.expenseUUID, params.descriptionExpenseUUID).
			Delete(&model.DescriptionExpenses{}).Error; errDelete != nil {
			r.Logger.Error("failed to delete expense: " + errDelete.Error())
			return errDelete
		}
		if errUpdate := tx.Model(&model.Expenses{}).
			Where("expense_uuid = ? AND budget_uuid = ?", params.expenseUUID, params.budgetUUID).
			Update(params.categoryExpense, gorm.Expr(params.categoryExpense+"::numeric - "+params.expense+" ::numeric")).
			Error; errUpdate != nil {
			r.Logger.Error("failed to update expense: " + errUpdate.Error())
			return errUpdate
		}
		return nil
	})
}
func (r *RepositoryExpense) ListDescriptionExpense(ctxRequest context.Context, expenseUUID string, limit, offset int) ([]model.DescriptionExpenses, error) {
	sliceDescriptionExpense := make([]model.DescriptionExpenses, 0, 50)
	if errList := r.Postgres.
		WithContext(ctxRequest).
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
func (r *RepositoryExpense) ExpenseExist(ctxRequest context.Context, budgetUUID, expenseUUID string) bool {
	var exist bool
	if errExist := r.Postgres.
		WithContext(ctxRequest).
		Raw(`SELECT EXISTS(
				 SELECT FROM expenses
				 WHERE budget_uuid = ? AND expense_uuid = ?)`, budgetUUID, expenseUUID).Scan(&exist).
		Error; errExist != nil {
		r.Logger.Error("failed to check the expense existence: " + errExist.Error())
		return false
	}
	return exist
}
