package expense

import (
	"app/budget-planner/internal/common"
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/model"

	"github.com/google/uuid"
)

type ServiceExpense struct {
	Repo    *RepositoryExpense
	IBudget di.IRepoBudget
}

func NewServiceExpense(repo *RepositoryExpense, repoBudget di.IRepoBudget) *ServiceExpense {
	return &ServiceExpense{
		Repo:    repo,
		IBudget: repoBudget,
	}
}
func (s *ServiceExpense) CreateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string) (*model.DescriptionExpense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	descriptionExpense := &model.DescriptionExpense{
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
			return nil, []string{ErrFailedCreateExpense.Error()}
		}
	} else {
		descriptionExpense.ExpenseUUID = expenseUUID
		descriptionExpense.DescriptionExpenseUUID = uuid.New().String()
		if errUpsert := s.Repo.UpsertExpense(descriptionExpense, budgetUUID, expenseUUID); errUpsert != nil {
			return nil, []string{ErrFailedCreateExpense.Error()}
		}
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) UpdateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string, descriptionExpenseUUID string) (*model.DescriptionExpense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, []string{custom_errors.ErrNotFoundExpense.Error()}
	}
	descriptionExpense, errGetDescriptionExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	var oldExpense string
	if errGetDescriptionExpense != nil {
		return nil, []string{ErrNotFoundDescriptionExpense.Error()}
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
			return nil, []string{ErrFailedUpdateExpense.Error()}
		}
	} else {
		if s.Repo.UpdateExpenseTransaction(descriptionExpense, oldExpense, budgetUUID, expenseUUID) != nil {
			return nil, []string{ErrFailedUpdateExpense.Error()}
		}
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) GetExpense(userUUID, budgetUUID, descriptionExpenseUUID string) (*model.DescriptionExpense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		return nil, []string{custom_errors.ErrNotFoundExpense.Error()}
	}
	descriptionExpense, errGetDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errGetDescExpense != nil {
		return nil, []string{ErrNotFoundDescriptionExpense.Error()}
	}
	return descriptionExpense, nil
}
func (s *ServiceExpense) DeleteExpense(userUUID, budgetUUID, descriptionExpenseUUID string) []string {
	sliceError := make([]string, 0, 3)
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		sliceError = append(sliceError, errValidate.Error())
	}
	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		sliceError = append(sliceError, custom_errors.ErrNotFoundExpense.Error())
	}
	descriptionExpense, errDescExpense := s.Repo.GetDescriptionExpense(expenseUUID, descriptionExpenseUUID)
	if errDescExpense != nil {
		sliceError = append(sliceError, ErrNotFoundDescriptionExpense.Error())
	}
	if len(sliceError) != 0 || errDescExpense != nil {
		return sliceError
	}
	if s.Repo.DeleteExpense(&deleteExpenseParams{
		categoryExpense:        descriptionExpense.Category,
		expense:                descriptionExpense.Expense,
		budgetUUID:             budgetUUID,
		expenseUUID:            expenseUUID,
		descriptionExpenseUUID: descriptionExpenseUUID,
	}) != nil {
		return []string{ErrFailedDeleteExpense.Error()}
	}
	return nil
}
func (s *ServiceExpense) ListExpense(budgetUUID, limitStr, offsetStr string) ([]model.DescriptionExpense, []string) {
	sliceError := make([]string, 0, 3)
	limit, offset, errPag := common.PaginationHelper(limitStr, offsetStr)
	if len(errPag) != 0 {
		sliceError = append(sliceError, errPag...)
	}

	expenseUUID, errGetExpense := s.Repo.GetExpenseUUID(budgetUUID)
	if errGetExpense != nil {
		sliceError = append(sliceError, custom_errors.ErrNotFoundExpense.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	descriptionExpenseList, errList := s.Repo.ListExpense(expenseUUID, limit, offset)
	if errList != nil {
		return nil, []string{ErrNotFoundDescriptionExpense.Error()}
	}
	return descriptionExpenseList, nil
}
func (s *ServiceExpense) helperValidateExpense(userUUID, budgetUUID string) error {
	if _, errBudgetUUID := uuid.Parse(budgetUUID); errBudgetUUID != nil {
		return custom_errors.ErrIncorrectFormatBudgetUUID
	}
	if !s.IBudget.BudgetExist(userUUID, budgetUUID) {
		return custom_errors.ErrNotFoundBudget
	}
	return nil
}
