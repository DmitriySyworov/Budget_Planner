package budget

import "errors"

var (
	ErrFailedCreateBudget = errors.New("failed to create budget")
	ErrOverlapStartFinish = errors.New("the start and finish parameters are overlap with another budget time period")
	ErrIncorrectStart     = errors.New("the start parameter must be in the format YYYY-MM-DD")
	ErrIncorrectFinish    = errors.New("the finish parameter must be in the format YYYY-MM-DD")
	ErrIncorrectDates     = errors.New("the start parameter cannot be greater than or equal to finish")
	ErrFailedRemoveBudget = errors.New("failed to remove budget")
	ErrFailedDeleteBudget = errors.New("failed to delete budget")
	ErrFailedUpdateBudget = errors.New("failed to update budget")
)
