package finance

import (
	"app/budget-planner/internal/apperrors"
	"app/budget-planner/internal/di"
	"context"
	"errors"
	"shared/sherrors"

	"github.com/google/uuid"
)

type ServiceFinance struct {
	Repo *RepositoryFinance
	di.IRepoExpense
	di.IRepoBudget
}

func NewServiceFinance(repoFinance *RepositoryFinance, repoBudget di.IRepoBudget, repoExpense di.IRepoExpense) *ServiceFinance {
	return &ServiceFinance{
		Repo:         repoFinance,
		IRepoBudget:  repoBudget,
		IRepoExpense: repoExpense,
	}
}

var ErrFailedGetFinance = errors.New("failed to get finance")

func (s *ServiceFinance) Finance(ctxRequest context.Context, userUUID, budgetUUID, expenseUUID string) (*Finance, error) {
	mapError := sherrors.MapError{Map: make(map[string]string, 2)}
	if _, errBudgetUUID := uuid.Parse(budgetUUID); errBudgetUUID != nil {
		mapError.Map["budget"] = apperrors.ErrIncorrectFormatBudgetUUID.Error()
	}
	if _, errExpenseUUID := uuid.Parse(expenseUUID); errExpenseUUID != nil {
		mapError.Map["expense"] = apperrors.ErrIncorrectFormatExpenseUUID.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	if !s.IRepoBudget.BudgetExist(ctxRequest, userUUID, budgetUUID) {
		return nil, apperrors.ErrNotFoundBudget
	}
	if !s.IRepoExpense.ExpenseExist(ctxRequest, budgetUUID, expenseUUID) {
		return nil, apperrors.ErrNotFoundExpense
	}
	dtoFinance, errGetFinance := s.Repo.Finance(ctxRequest, budgetUUID, expenseUUID)
	if errGetFinance != nil {
		return nil, ErrFailedGetFinance
	}
	return &Finance{
		&Budget{
			BudgetInitial:               dtoFinance.BudgetInitial,
			BudgetBalance:               dtoFinance.BudgetBalance,
			PredictedAverageSpendPerDay: dtoFinance.PredictedAverageSpendPerDay,
		},
		&Expenses{
			Health:      dtoFinance.Health,
			Sport:       dtoFinance.Sport,
			Supermarket: dtoFinance.Supermarket,
			Restaurant:  dtoFinance.Restaurant,
			Other:       dtoFinance.Other,
			Savings:     dtoFinance.Savings,
			Investments: dtoFinance.Investments,
			Leisure:     dtoFinance.Leisure,
		},
		&ExpensesPercent{
			HealthExpensePercent:      dtoFinance.HealthExpensePercent,
			SportExpensePercent:       dtoFinance.SportExpensePercent,
			SupermarketExpensePercent: dtoFinance.SupermarketExpensePercent,
			RestaurantExpensePercent:  dtoFinance.RestaurantExpensePercent,
			OtherExpensePercent:       dtoFinance.OtherExpensePercent,
			SavingsExpensePercent:     dtoFinance.SavingsExpensePercent,
			InvestmentsExpensePercent: dtoFinance.InvestmentsExpensePercent,
			LeisureExpensePercent:     dtoFinance.LeisureExpensePercent,
		},
	}, nil
}
