package expense

import "errors"

var (
	ErrExpenseAlreadyExist       = errors.New("this expense already exist")
	ErrNotFoundExpense           = errors.New("not found expense")
	ErrIncorrectFormatBudgetUUID = errors.New("the budget uuid must be exactly 36 characters")
	ErrFailedRemoveExpense       = errors.New("failed to remove expense")
	ErrFailedDeleteExpense       = errors.New("failed to delete expense")
	ErrFailedCreateExpense       = errors.New("failed to create expense")
	ErrFailedUpdateExpense       = errors.New("failed to update expense")
)
