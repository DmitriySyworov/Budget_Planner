package expense

import "errors"

var (
	ErrNotFoundDescriptionExpense = errors.New("not found description expense")
	ErrIncorrectDescription       = errors.New("description cannot be more than 250 characters")
	ErrIncorrectCategory          = errors.New("category mast be a health, sport, supermarket, restaurant, leisure, investments, savings or other")
	ErrFailedDeleteExpense        = errors.New("failed to delete expense")
	ErrFailedCreateExpense        = errors.New("failed to create expense")
	ErrFailedUpdateExpense        = errors.New("failed to update expense")
)
