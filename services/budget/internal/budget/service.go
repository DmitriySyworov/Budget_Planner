package budget

import (
	"app/budget-planner/internal/apperrors"
	"app/budget-planner/internal/model"
	"errors"
	"shared/loggers"
	"shared/pagination"
	"shared/shconstant"
	"shared/sherrors"
	"shared/shprotos/event"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type ServiceBudget struct {
	Repo   IRepositoryBudget
	Logger *loggers.Logger
}

func NewServiceBudget(repo IRepositoryBudget, logger *loggers.Logger) *ServiceBudget {
	return &ServiceBudget{
		Repo:   repo,
		Logger: logger,
	}
}
func (s *ServiceBudget) CreateBudget(body *RequestCreateBudget, userUUID string) (*model.Budgets, error) {
	mapError := sherrors.MapError{Map: make(map[string]string, 3)}
	start, errStart := time.Parse(time.DateOnly, body.Start)
	if errStart != nil {
		mapError.Map["start"] = ErrIncorrectStart.Error()
	}
	finish, errFinish := time.Parse(time.DateOnly, body.Finish)
	if errFinish != nil {
		mapError.Map["finish"] = ErrIncorrectFinish.Error()
	}
	if errFinish == nil && errStart == nil && start.Unix() >= finish.Unix() {
		mapError.Map["dates"] = ErrIncorrectDates.Error()
	}
	if s.Repo.DateOverlapCreate(userUUID, start, finish) {
		mapError.Map["dates"] = ErrOverlapStartFinish.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	budget := &model.Budgets{
		Amount:      body.Amount,
		Start:       start,
		Finish:      finish,
		Description: body.Description,
		BudgetUUID:  uuid.New().String(),
		UserUUID:    userUUID,
	}
	if errCreate := s.Repo.CreateBudget(budget); errCreate != nil {
		return nil, ErrFailedCreateBudget
	}
	return budget, nil
}
func (s *ServiceBudget) UpdateBudget(body *RequestUpdateBudget, userUUID, budgetUUID string) (*model.Budgets, error) {
	mapError := sherrors.MapError{Map: make(map[string]string, 4)}
	budget, errValidate := s.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		mapError.Map["budget"] = errValidate.Error()
	}
	var start, finish time.Time
	var errStart, errFinish error
	if body.Start != "" {
		start, errStart = time.Parse(time.DateOnly, body.Start)
		if errStart != nil {
			mapError.Map["start"] = ErrIncorrectStart.Error()
		}
	}
	if body.Finish != "" {
		finish, errFinish = time.Parse(time.DateOnly, body.Finish)
		if errFinish != nil {
			mapError.Map["finish"] = ErrIncorrectFinish.Error()
		}
	}
	if body.Start != "" && body.Finish != "" && errStart == nil && errFinish == nil {
		if start.Unix() >= finish.Unix() {
			mapError.Map["dates"] = ErrIncorrectDates.Error()
		}
		if s.Repo.DateOverlapUpdate(userUUID, budgetUUID, start, finish) {
			mapError.Map["dates"] = ErrOverlapStartFinish.Error()
		}
	} else if body.Start != "" && body.Finish == "" && errStart == nil && errValidate == nil {
		if start.Unix() >= budget.Finish.Unix() {
			mapError.Map["dates"] = ErrIncorrectDates.Error()
		}
		if s.Repo.DateOverlapUpdate(userUUID, budgetUUID, start, budget.Finish) {
			mapError.Map["dates"] = ErrOverlapStartFinish.Error()
		}
	} else if body.Start == "" && body.Finish != "" && errFinish == nil && errValidate == nil {
		if budget.Start.Unix() >= finish.Unix() {
			mapError.Map["dates"] = ErrIncorrectDates.Error()
		}
		if s.Repo.DateOverlapUpdate(userUUID, budgetUUID, budget.Start, finish) {
			mapError.Map["dates"] = ErrOverlapStartFinish.Error()
		}
	}
	if len(mapError.Map) != 0 || errValidate != nil {
		return nil, mapError
	}
	budget.Amount = body.Amount
	budget.Start = start
	budget.Finish = finish
	budget.Description = body.Description
	errUpdate := s.Repo.UpdateBudget(budget, userUUID, budgetUUID)
	if errUpdate != nil {
		return nil, ErrFailedUpdateBudget
	}
	return budget, nil
}
func (s *ServiceBudget) GetBudget(userUUID, budgetUUID string) (*model.Budgets, error) {
	budget, errValidate := s.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		return nil, errValidate
	}
	return budget, nil
}

func (s *ServiceBudget) RemoveBudget(userUUID, budgetUUID, typeRemove string) error {
	mapError := sherrors.MapError{Map: make(map[string]string, 2)}
	_, errValidate := s.HelperValidateBudget(userUUID, budgetUUID)
	if errValidate != nil {
		mapError.Map["budget"] = errValidate.Error()
	}
	if typeRemove != shconstant.TypeSoftDelete && typeRemove != shconstant.TypeHardDelete && typeRemove != "" {
		mapError.Map["type"] = sherrors.ErrIncorrectTypeRemove.Error()
	}

	if len(mapError.Map) != 0 {
		return mapError
	}
	if typeRemove == shconstant.TypeSoftDelete || typeRemove == "" {
		errRemove := s.Repo.RemoveBudget(userUUID, budgetUUID)
		if errRemove != nil {
			return ErrFailedRemoveBudget
		}
	} else if typeRemove == shconstant.TypeHardDelete {
		errDelete := s.Repo.DeleteBudget(userUUID, budgetUUID)
		if errDelete != nil {
			return ErrFailedDeleteBudget
		}
	} else {
		return sherrors.ErrIncorrectTypeRemove
	}
	return nil
}
func (s *ServiceBudget) HelperValidateBudget(userUUID, budgetUUID string) (*model.Budgets, error) {
	if _, errBudgetUUID := uuid.Parse(budgetUUID); errBudgetUUID != nil {
		return nil, apperrors.ErrIncorrectFormatBudgetUUID
	}
	budget, errGetBudget := s.Repo.GetBudget(userUUID, budgetUUID)
	if errGetBudget != nil {
		return nil, apperrors.ErrNotFoundBudget
	}
	return budget, nil
}

func (s *ServiceBudget) ListBudget(userUUID, limitStr, offsetStr string) ([]model.Budgets, error) {
	mapError := &sherrors.MapError{Map: make(map[string]string, 2)}
	limit, offset := pagination.HelperPagination(limitStr, offsetStr, mapError)
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	listBudget, errList := s.Repo.ListBudget(userUUID, limit, offset)
	if errList != nil {
		return nil, apperrors.ErrNotFoundBudget
	}
	return listBudget, nil
}
func (s *ServiceBudget) DeleteDataDeletingUser(data []byte) error {
	var eventDel event.DeleteUserDataEvent
	if errUnmarshal := proto.Unmarshal(data, &eventDel); errUnmarshal != nil {
		s.Logger.Error("failed to unmarshal delete event: " + errUnmarshal.Error())
		return errUnmarshal
	}
	userUUID := eventDel.GetUserUuid()
	if userUUID == "" {
		s.Logger.Error("event with an empty user_uuid was sent")
		return errors.New("event with an empty user_uuid was sent")
	}
	if errDel := s.Repo.DeleteAllUserBudgets(userUUID); errDel != nil {
		s.Logger.Warn("failed to delete user budgets maybe it was deleted earlier: " + errDel.Error())
		return errDel
	}
	s.Logger.Info("cascade deletion of all user data was successful")
	return nil
}
