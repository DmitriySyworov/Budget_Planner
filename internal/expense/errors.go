package expense

import "errors"

var (
	ErrNotFoundDescriptionExpense = errors.New("not found description expense")
	ErrNotFoundExpense            = errors.New("not found expense")
	ErrIncorrectFormatBudgetUUID  = errors.New("the budget uuid must be exactly 36 characters")
	ErrIncorrectDescription       = errors.New("description cannot be more than 250 characters")
	ErrIncorrectCategory          = errors.New("category mast be a health, sport, supermarket, restaurant, leisure, investments, savings or other")
	ErrFailedRemoveExpense        = errors.New("failed to remove expense")
	ErrFailedDeleteExpense        = errors.New("failed to delete expense")
	ErrFailedCreateExpense        = errors.New("failed to create expense")
	ErrFailedUpdateExpense        = errors.New("failed to update expense")
)
