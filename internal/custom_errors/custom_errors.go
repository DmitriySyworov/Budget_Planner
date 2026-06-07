package custom_errors

import "errors"

var (
	ErrIncorrectToken               = errors.New("incorrect token")
	ErrNotFoundUser                 = errors.New("such user does not exist")
	ErrFailedAssertionContextValues = errors.New("failed to assert type ContextValues: ")
	ErrIncorrectFormatUserUUID      = errors.New("the user uuid must be exactly 36 characters")
	ErrNotFoundBudget               = errors.New("not found budget")
	ErrIncorrectTypeRemove          = errors.New("the type  must be a soft-delete ot hard-delete")
	ErrIncorrectDecimal             = errors.New(" must be a positive decimals and greater than 0")
)
