package budget

import (
	"app/budget-planner/internal/common"
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/model"
	"time"

	"github.com/google/uuid"
)

type ServiceBudget struct {
	Repo *RepositoryBudget
	di.IRepoUser
}

func NewServiceBudget(repo *RepositoryBudget, repoUser di.IRepoUser) *ServiceBudget {
	return &ServiceBudget{
		Repo:      repo,
		IRepoUser: repoUser,
	}
}
func (s *ServiceBudget) CreateBudget(body *CreateAndUpdateBudget, userUUID string) (*model.Budget, []string) {
	sliceError := make([]string, 0, 4)
	if !s.IRepoUser.IsUserExistsByUUID(userUUID) {
		sliceError = append(sliceError, custom_errors.ErrNotFoundUser.Error())
	}
	start, errStart := time.Parse(time.DateOnly, body.Start)
	if errStart != nil {
		sliceError = append(sliceError, ErrIncorrectStart.Error())
	}
	finish, errFinish := time.Parse(time.DateOnly, body.Finish)
	if errFinish != nil {
		sliceError = append(sliceError, ErrIncorrectFinish.Error())
	}

	if s.Repo.DateOverlap(start, finish) {
		sliceError = append(sliceError, ErrOverlapStartFinish.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	budget := &model.Budget{
		Amount:      body.Amount,
		Start:       start,
		Finish:      finish,
		Description: body.Description,
		BudgetUUID:  uuid.New().String(),
		UserUUID:    userUUID,
	}
	if errCreate := s.Repo.Create(budget); errCreate != nil {
		return nil, []string{ErrFailedCreateBudget.Error()}
	}
	return budget, nil
}
func (s *ServiceBudget) UpdateBudget(body *CreateAndUpdateBudget, userUUID, budgetUUID string) (*model.Budget, []string) {
	sliceError := make([]string, 0, 5)
	budget, errValidate := s.helperValidateBudget(userUUID, budgetUUID)
	sliceError = append(sliceError, errValidate...)
	var start, finish time.Time
	var errStart, errFinish error
	if body.Start != "" {
		start, errStart = time.Parse(time.DateOnly, body.Start)
		if errStart != nil {
			sliceError = append(sliceError, ErrIncorrectStart.Error())
		}
	}
	if body.Finish != "" {
		finish, errFinish = time.Parse(time.DateOnly, body.Finish)
		if errFinish != nil {
			sliceError = append(sliceError, ErrIncorrectFinish.Error())
		}
	}
	if body.Start != "" && body.Finish != "" && errStart == nil && errFinish == nil {
		if s.Repo.DateOverlap(start, finish) {
			sliceError = append(sliceError, ErrOverlapStartFinish.Error())
		}
	} else if body.Start != "" && body.Finish == "" && errStart == nil {
		if s.Repo.DateOverlap(start, budget.Finish) {
			sliceError = append(sliceError, ErrOverlapStartFinish.Error())
		}
	} else if body.Start == "" && body.Finish != "" && errFinish == nil {
		if s.Repo.DateOverlap(budget.Start, finish) {
			sliceError = append(sliceError, ErrOverlapStartFinish.Error())
		}
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	budget.Amount = body.Amount
	budget.Start = start
	budget.Finish = finish
	budget.Description = body.Description
	errUpdate := s.Repo.UpdateBudget(budget, userUUID, budgetUUID)
	if errUpdate != nil {
		return nil, []string{ErrFailedUpdateBudget.Error()}
	}
	return budget, nil
}
func (s *ServiceBudget) GetBudget(userUUID, budgetUUID string) (*model.Budget, []string) {
	budget, errValidate := s.helperValidateBudget(userUUID, budgetUUID)
	if len(errValidate) != 0 {
		return nil, errValidate
	}
	return budget, nil
}

func (s *ServiceBudget) RemoveBudget(userUUID, budgetUUID, typeRemove string) []string {
	sliceError := make([]string, 0, 4)
	_, errValidate := s.helperValidateBudget(userUUID, budgetUUID)
	sliceError = append(sliceError, errValidate...)
	if typeRemove != common.TypeSoftDelete && typeRemove != common.TypeHardDelete && typeRemove != "" {
		sliceError = append(sliceError, custom_errors.ErrIncorrectTypeRemove.Error())
	}
	if len(sliceError) != 0 {
		return sliceError
	}
	if typeRemove == common.TypeSoftDelete || typeRemove == "" {
		errRemove := s.Repo.RemoveBudget(userUUID, budgetUUID)
		if errRemove != nil {
			return []string{ErrFailedRemoveBudget.Error()}
		}
	} else if typeRemove == common.TypeHardDelete {
		errDelete := s.Repo.DeleteBudget(userUUID, budgetUUID)
		if errDelete != nil {
			return []string{ErrFailedDeleteBudget.Error()}
		}
	}
	return []string{custom_errors.ErrIncorrectTypeRemove.Error()}
}
func (s *ServiceBudget) helperValidateBudget(userUUID, budgetUUID string) (*model.Budget, []string) {
	sliceError := make([]string, 0, 3)
	if !s.IRepoUser.IsUserExistsByUUID(userUUID) {
		sliceError = append(sliceError, custom_errors.ErrNotFoundUser.Error())
	}
	budget, errGetBudget := s.Repo.GetBudget(userUUID, budgetUUID)
	if errGetBudget != nil {
		sliceError = append(sliceError, custom_errors.ErrNotFoundBudget.Error())
	}
	return budget, sliceError
}

func (s *ServiceBudget) ListBudget(userUUID, limitStr, offsetStr string) ([]model.Budget, []string) {
	sliceError := make([]string, 0, 3)
	if !s.IRepoUser.IsUserExistsByUUID(userUUID) {
		sliceError = append(sliceError, custom_errors.ErrNotFoundUser.Error())
	}
	limit, offset, errPagination := common.PaginationHelper(limitStr, offsetStr)
	sliceError = append(sliceError, errPagination...)
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	listBudget, errList := s.Repo.ListBudget(userUUID, limit, offset)
	if errList != nil {
		return nil, []string{custom_errors.ErrNotFoundBudget.Error()}
	}
	return listBudget, nil
}
