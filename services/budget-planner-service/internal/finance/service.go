package finance

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"errors"
	"shared/shared_errors"

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

func (s *ServiceFinance) Finance(userUUID, budgetUUID, expenseUUID string) (*Finance, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
	if _, errBudgetUUID := uuid.Parse(budgetUUID); errBudgetUUID != nil {
		mapError.Map["budget"] = custom_errors.ErrIncorrectFormatBudgetUUID.Error()
	}
	if _, errExpenseUUID := uuid.Parse(expenseUUID); errExpenseUUID != nil {
		mapError.Map["expense"] = custom_errors.ErrIncorrectFormatExpenseUUID.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	if !s.IRepoBudget.BudgetExist(userUUID, budgetUUID) {
		return nil, custom_errors.ErrNotFoundBudget
	}
	if !s.IRepoExpense.ExpenseExist(budgetUUID, expenseUUID) {
		return nil, custom_errors.ErrNotFoundExpense
	}
	dtoFinance, errGetFinance := s.Repo.Finance(budgetUUID, expenseUUID)
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
