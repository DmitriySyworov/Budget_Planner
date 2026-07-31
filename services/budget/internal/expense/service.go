package expense

import (
	"app/budget-planner/internal/apperrors"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/model"
	"shared/pagination"
	"shared/sherrors"

	"github.com/google/uuid"
)

type ServiceExpense struct {
	Repo           IRepositoryExpense
	IServiceBudget di.IServiceBudget
}

func NewServiceExpense(repo IRepositoryExpense, serviceBudget di.IServiceBudget) *ServiceExpense {
	return &ServiceExpense{
		Repo:           repo,
		IServiceBudget: serviceBudget,
	}
}
func (s *ServiceExpense) CreateExpense(body *RequestCreateDescriptionExpense, userUUID, budgetUUID string) (*ResponseCreateAndUpdateExpense, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	descriptionExpense := &model.DescriptionExpenses{
		Category:    body.Category,
		Expense:     body.Expense,
		Description: body.Description,
	}
	expenseUUID, errGetExpenseUUID := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpenseUUID != nil {
		newExpenseUUID := uuid.New().String()
		descriptionExpense.ExpenseUUID = newExpenseUUID
		descriptionExpense.DescriptionExpenseUUID = uuid.New().String()
		if errUpsert := s.Repo.UpsertExpense(descriptionExpense, budgetUUID, newExpenseUUID); errUpsert != nil {
			return nil, ErrFailedCreateExpense
		}
	} else {
		descriptionExpense.ExpenseUUID = expenseUUID
		descriptionExpense.DescriptionExpenseUUID = uuid.New().String()
		if errUpsert := s.Repo.UpsertExpense(descriptionExpense, budgetUUID, expenseUUID); errUpsert != nil {
			return nil, ErrFailedCreateExpense
		}
	}
	expense, errGetExpense := s.Repo.GetExpense(budgetUUID)
	if errGetExpense != nil {
		return nil, apperrors.ErrNotFoundExpense
	}
	return &ResponseCreateAndUpdateExpense{
		Expenses:            expense,
		DescriptionExpenses: descriptionExpense,
	}, nil
}
func (s *ServiceExpense) UpdateExpense(body *RequestUpdateDescriptionExpense, userUUID, budgetUUID string, descriptionExpenseUUID string) (*ResponseCreateAndUpdateExpense, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, apperrors.ErrNotFoundExpense
	}
	descriptionExpense, errGetDescriptionExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errGetDescriptionExpense != nil {
		return nil, ErrNotFoundDescriptionExpense
	}
	if body.Description != "" && body.Expense == "" && body.Category == "" {
		if s.Repo.UpdateDescriptionExpense(descriptionExpense) != nil {
			return nil, ErrFailedUpdateExpense
		}
	} else {
		oldExpense := descriptionExpense.Expense
		oldCategory := descriptionExpense.Category
		descriptionExpense.Category = body.Category
		descriptionExpense.Description = body.Description
		descriptionExpense.Expense = body.Expense
		if s.Repo.UpdateExpenseTransaction(descriptionExpense, oldExpense, oldCategory, budgetUUID, expenseUUID) != nil {
			return nil, ErrFailedUpdateExpense
		}
	}
	expense, errGetExpense := s.Repo.GetExpense(budgetUUID)
	if errGetExpense != nil {
		return nil, apperrors.ErrNotFoundExpense
	}
	return &ResponseCreateAndUpdateExpense{
		Expenses:            expense,
		DescriptionExpenses: descriptionExpense,
	}, nil
}
func (s *ServiceExpense) GetDescriptionExpense(userUUID, budgetUUID, descriptionExpenseUUID string) (*model.DescriptionExpenses, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, apperrors.ErrNotFoundExpense
	}
	descriptionExpense, errGetDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errGetDescExpense != nil {
		return nil, ErrNotFoundDescriptionExpense
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) DeleteDescriptionExpense(userUUID, budgetUUID, descriptionExpenseUUID string) error {
	mapError := sherrors.MapError{Map: make(map[string]string, 3)}
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		mapError.Map["budget"] = errValidate.Error()
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		mapError.Map["expense"] = apperrors.ErrNotFoundExpense.Error()
	}
	descriptionExpense, errDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errDescExpense != nil {
		mapError.Map["expense"] = ErrNotFoundDescriptionExpense.Error()
	}
	if len(mapError.Map) != 0 || errDescExpense != nil {
		return mapError
	}
	if s.Repo.DeleteDescriptionExpense(&deleteExpenseParams{
		categoryExpense:        descriptionExpense.Category,
		expense:                descriptionExpense.Expense,
		budgetUUID:             budgetUUID,
		expenseUUID:            expenseUUID,
		descriptionExpenseUUID: descriptionExpenseUUID,
	}) != nil {
		return ErrFailedDeleteExpense
	}
	return nil
}
func (s *ServiceExpense) ListDescriptionExpense(budgetUUID, limitStr, offsetStr string) ([]model.DescriptionExpenses, error) {
	mapError := &sherrors.MapError{Map: make(map[string]string, 3)}
	limit, offset := pagination.HelperPagination(limitStr, offsetStr, mapError)
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		mapError.Map["expense"] = apperrors.ErrNotFoundExpense.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	descriptionExpenseList, errList := s.Repo.ListDescriptionExpense(expenseUUID, limit, offset)
	if errList != nil {
		return nil, ErrNotFoundDescriptionExpense
	}
	return descriptionExpenseList, nil
}
func (s *ServiceExpense) GetExpense(budgetUUID string) (*model.Expenses, error) {
	if _, errUUID := uuid.Parse(budgetUUID); errUUID != nil {
		return nil, apperrors.ErrIncorrectFormatBudgetUUID
	}
	expense, errGetExpense := s.Repo.GetExpense(budgetUUID)
	if errGetExpense != nil {
		return nil, apperrors.ErrNotFoundExpense
	}
	return expense, nil
}
