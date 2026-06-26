package expense

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/model"
	"errors"
	"shared/shared_common"
	"shared/shared_errors"

	"github.com/google/uuid"
)

type ServiceExpense struct {
	Repo           *RepositoryExpense
	IServiceBudget di.IServiceBudget
}

func NewServiceExpense(repo *RepositoryExpense, serviceBudget di.IServiceBudget) *ServiceExpense {
	return &ServiceExpense{
		Repo:           repo,
		IServiceBudget: serviceBudget,
	}
}
func (s *ServiceExpense) CreateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string) (*model.DescriptionExpenses, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	descriptionExpense := &model.DescriptionExpenses{
		Category:    body.Category,
		Expense:     body.Expense,
		Description: body.Description,
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		newExpenseUUID := uuid.New().String()
		descriptionExpense.ExpenseUUID = newExpenseUUID
		descriptionExpense.DescriptionExpenseUUID = uuid.New().String()
		if errUpsert := s.Repo.UpsertExpense(descriptionExpense, budgetUUID, expenseUUID); errUpsert != nil {
			return nil, ErrFailedCreateExpense
		}
	} else {
		descriptionExpense.ExpenseUUID = expenseUUID
		descriptionExpense.DescriptionExpenseUUID = uuid.New().String()
		if errUpsert := s.Repo.UpsertExpense(descriptionExpense, budgetUUID, expenseUUID); errUpsert != nil {
			return nil, ErrFailedCreateExpense
		}
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) UpdateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string, descriptionExpenseUUID string) (*model.DescriptionExpenses, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, custom_errors.ErrNotFoundExpense
	}
	descriptionExpense, errGetDescriptionExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	var oldExpense string
	if errGetDescriptionExpense != nil {
		return nil, ErrNotFoundDescriptionExpense
	} else {
		oldExpense = descriptionExpense.Expense
	}
	if body.Category != "" {
		descriptionExpense.Category = body.Category
	}
	if body.Description != "" {
		descriptionExpense.Description = body.Description
	}
	if body.Expense != "" {
		descriptionExpense.Expense = body.Expense
		if s.Repo.UpdateDescriptionExpense(descriptionExpense, expenseUUID, descriptionExpenseUUID) != nil {
			return nil, ErrFailedUpdateExpense
		}
	} else {
		if s.Repo.UpdateExpenseTransaction(descriptionExpense, oldExpense, budgetUUID, expenseUUID) != nil {
			return nil, ErrFailedUpdateExpense
		}
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) GetExpense(userUUID, budgetUUID, descriptionExpenseUUID string) (*model.DescriptionExpenses, error) {
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, custom_errors.ErrNotFoundExpense
	}
	descriptionExpense, errGetDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errGetDescExpense != nil {
		return nil, ErrNotFoundDescriptionExpense
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) DeleteExpense(userUUID, budgetUUID, descriptionExpenseUUID string) error {
	mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
	_, errValidate := s.IServiceBudget.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		mapError.Map["budget"] = errValidate.Error()
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		mapError.Map["expense"] = custom_errors.ErrNotFoundExpense.Error()
	}
	descriptionExpense, errDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errDescExpense != nil {
		mapError.Map["expense"] = ErrNotFoundDescriptionExpense.Error()
	}
	if len(mapError.Map) != 0 || errDescExpense != nil {
		return mapError
	}
	if s.Repo.DeleteExpense(&deleteExpenseParams{
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
func (s *ServiceExpense) ListExpense(budgetUUID, limitStr, offsetStr string) ([]model.DescriptionExpenses, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
	limit, offset, errPagination := shared_common.PaginationHelper(limitStr, offsetStr)
	if len(errPagination) != 0 {
		for _, err := range errPagination {
			switch {
			case errors.Is(err, shared_errors.ErrIncorrectLimit):
				mapError.Map["limit"] = shared_errors.ErrIncorrectLimit.Error()
			case errors.Is(err, shared_errors.ErrIncorrectOffset):
				mapError.Map["offset"] = shared_errors.ErrIncorrectOffset.Error()
			}
		}
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		mapError.Map["expense"] = custom_errors.ErrNotFoundExpense.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	descriptionExpenseList, errList := s.Repo.ListExpense(expenseUUID, limit, offset)
	if errList != nil {
		return nil, ErrNotFoundDescriptionExpense
	}
	return descriptionExpenseList, nil
}
