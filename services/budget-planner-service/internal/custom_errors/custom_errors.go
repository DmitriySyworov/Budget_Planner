package custom_errors

import "errors"

var (
	ErrNotFoundExpense            = errors.New("not found expense")
	ErrIncorrectFormatBudgetUUID  = errors.New("the budget uuid must be exactly 36 characters")
	ErrIncorrectFormatExpenseUUID = errors.New("the expense uuid must be exactly 36 characters")
	ErrNotFoundBudget             = errors.New("not found budget")
	ErrIncorrectDecimal           = errors.New(" must be a positive decimals and greater than 0")
)
