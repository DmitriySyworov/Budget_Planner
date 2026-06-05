package budget

import "errors"

var (
	ErrFailedCreateBudget  = errors.New("failed to create budget")
	ErrNotFoundBudget      = errors.New("not found budget")
	ErrOverlapStartFinish  = errors.New("the start and finish parameters are overlap with another budget time period")
	ErrIncorrectStart      = errors.New("the start parameter must be in the format YYYY-MM-DD")
	ErrIncorrectFinish     = errors.New("the finish parameter must be in the format YYYY-MM-DD")
	ErrIncorrectFormatUUID = errors.New("the budget uuid must be exactly 36 characters")
	ErrIncorrectTypeRemove = errors.New("the type  must be a soft-delete ot hard-delete")
	ErrFailedRemoveBudget  = errors.New("failed to remove budget")
	ErrFailedDeleteBudget  = errors.New("failed to delete budget")
	ErrFailedUpdateBudget  = errors.New("failed to update budget")
	ErrIncorrectAmount     = errors.New("amount must be a positive integer and greater than 0")
)
