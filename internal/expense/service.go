package expense

import (
	"app/budget-planner/internal/common"
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/model"
)

type ServiceExpense struct {
	Repo    *RepositoryExpense
	IUser   di.IRepoUser
	IBudget di.IRepoBudget
}

func NewServiceExpense(repo *RepositoryExpense, repoUser di.IRepoUser, repoBudget di.IRepoBudget) *ServiceExpense {
	return &ServiceExpense{
		Repo:    repo,
		IUser:   repoUser,
		IBudget: repoBudget,
	}
}
func (s *ServiceExpense) CreateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string) (*model.Expense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	if s.Repo.ExpenseExist(userUUID, budgetUUID) {
		return nil, []string{ErrExpenseAlreadyExist.Error()}
	}
	expense := &model.Expense{
		Health:      body.Health,
		Sport:       body.Sport,
		Supermarket: body.Supermarket,
		Restaurant:  body.Restaurant,
		Leisure:     body.Leisure,
		Investments: body.Investments,
		Savings:     body.Savings,
		Other:       body.Other,
		BudgetUUID:  budgetUUID,
		UserUUID:    userUUID,
	}
	if errCreate := s.Repo.CreateExpense(expense); errCreate != nil {
		return nil, []string{ErrFailedCreateExpense.Error()}
	}
	return expense, nil
}
func (s *ServiceExpense) UpdateExpense(body *CreateAndUpdateExpense, userUUID, budgetUUID string) (*model.Expense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	expense, errGet := s.Repo.GetExpense(userUUID, budgetUUID)
	if errGet != nil {
		return nil, []string{ErrNotFoundExpense.Error()}
	}
	if body.Sport != "" {
		expense.Sport = body.Sport
	}
	if body.Savings != "" {
		expense.Sport = body.Savings
	}
	if body.Leisure != "" {
		expense.Sport = body.Leisure
	}
	if body.Investments != "" {
		expense.Sport = body.Investments
	}
	if body.Restaurant != "" {
		expense.Sport = body.Restaurant
	}
	if body.Health != "" {
		expense.Sport = body.Health
	}
	if body.Other != "" {
		expense.Sport = body.Other
	}
	if body.Supermarket != "" {
		expense.Sport = body.Supermarket
	}
	if errUpdate := s.Repo.UpdateExpense(expense, userUUID, budgetUUID); errUpdate != nil {
		return nil, []string{ErrFailedUpdateExpense.Error()}
	}
	return expense, nil
}
func (s *ServiceExpense) GetExpense(userUUID, budgetUUID string) (*model.Expense, []string) {
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, []string{errValidate.Error()}
	}
	expense, errGet := s.Repo.GetExpense(userUUID, budgetUUID)
	if errGet != nil {
		return nil, []string{ErrNotFoundExpense.Error()}
	}
	return expense, nil
}
func (s *ServiceExpense) RemoveExpense(userUUID, budgetUUID, typeRemove string) []string {
	sliceError := make([]string, 0, 2)
	errValidate := s.helperValidateExpense(userUUID, budgetUUID)
	if errValidate != nil {
		sliceError = append(sliceError, errValidate.Error())
	}
	if typeRemove != common.TypeSoftDelete && typeRemove != common.TypeHardDelete && typeRemove != "" {
		sliceError = append(sliceError, custom_errors.ErrIncorrectTypeRemove.Error())
	}
	if len(sliceError) != 0 {
		return sliceError
	}
	if !s.Repo.ExpenseExist(userUUID, budgetUUID) {
		sliceError = append(sliceError, ErrNotFoundExpense.Error())
	}
	if len(sliceError) != 0 {
		return sliceError
	}
	if typeRemove == common.TypeSoftDelete || typeRemove == "" {
		errRemove := s.Repo.RemoveExpense(userUUID, budgetUUID)
		if errRemove != nil {
			return []string{ErrFailedRemoveExpense.Error()}
		}
	} else if typeRemove == common.TypeHardDelete {
		errDelete := s.Repo.DeleteExpense(userUUID, budgetUUID)
		if errDelete != nil {
			return []string{ErrFailedDeleteExpense.Error()}
		}
	}
	return []string{custom_errors.ErrIncorrectTypeRemove.Error()}
}
func (s *ServiceExpense) ListExpense(userUUID, limitStr, offsetStr string) ([]model.Expense, []string) {
	sliceError := make([]string, 0, 3)
	if !s.IUser.IsUserExistsByUUID(userUUID) {
		sliceError = append(sliceError, custom_errors.ErrFailedAssertionContextValues.Error())
	}
	limit, offset, errPag := common.PaginationHelper(limitStr, offsetStr)
	if len(errPag) != 0 {
		sliceError = append(sliceError, errPag...)
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	expenseList, errList := s.Repo.ListExpense(userUUID, limit, offset)
	if errList != nil {
		return nil, []string{ErrNotFoundExpense.Error()}
	}
	return expenseList, nil
}
func (s *ServiceExpense) helperValidateExpense(userUUID, budgetUUID string) error {
	if len(budgetUUID) != 36 {
		return ErrIncorrectFormatBudgetUUID
	}
	if !s.IUser.IsUserExistsByUUID(userUUID) {
		return custom_errors.ErrNotFoundUser
	}
	if !s.IBudget.BudgetExist(userUUID, budgetUUID) {
		return custom_errors.ErrNotFoundBudget
	}
	return nil
}
