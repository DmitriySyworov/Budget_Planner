package finance

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/di"
	"errors"

	"github.com/google/uuid"
)

type ServiceFinance struct {
	Repo *RepositoryFinance
	di.IRepoExpense
	di.IRepoBudget
	di.IRepoUser
}

func NewServiceFinance(repoFinance *RepositoryFinance, repoUser di.IRepoUser, repoBudget di.IRepoBudget, repoExpense di.IRepoExpense) *ServiceFinance {
	return &ServiceFinance{
		Repo:         repoFinance,
		IRepoUser:    repoUser,
		IRepoBudget:  repoBudget,
		IRepoExpense: repoExpense,
	}
}

var ErrFailedGetFinance = errors.New("failed to get finance")

func (s *ServiceFinance) Finance(userUUID, budgetUUID, expenseUUID string) (*Finance, []string) {
	sliceError := make([]string, 2)
	if _, errBudgetUUID := uuid.Parse(budgetUUID); errBudgetUUID != nil {
		sliceError = append(sliceError, custom_errors.ErrIncorrectFormatBudgetUUID.Error())
	}
	if _, errExpenseUUID := uuid.Parse(expenseUUID); errExpenseUUID != nil {
		sliceError = append(sliceError, custom_errors.ErrIncorrectFormatExpenseUUID.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	if !s.IRepoUser.IsUserExistsByUUID(userUUID) {
		return nil, []string{custom_errors.ErrNotFoundUser.Error()}
	}
	if !s.IRepoBudget.BudgetExist(userUUID, budgetUUID) {
		return nil, []string{custom_errors.ErrNotFoundBudget.Error()}
	}
	if !s.IRepoExpense.ExpenseExist(budgetUUID, expenseUUID) {
		return nil, []string{custom_errors.ErrNotFoundExpense.Error()}
	}
	dtoFinance, errGetFinance := s.Repo.Finance(budgetUUID, expenseUUID)
	if errGetFinance != nil {
		return nil, []string{ErrFailedGetFinance.Error()}
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
