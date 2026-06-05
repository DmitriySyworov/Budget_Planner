package custom_errors

import "errors"

var (
	ErrIncorrectToken               = errors.New("incorrect token")
	ErrNotFoundUser                 = errors.New("such user does not exist")
	ErrFailedAssertionContextValues = errors.New("failed to assert type ContextValues: ")
)
