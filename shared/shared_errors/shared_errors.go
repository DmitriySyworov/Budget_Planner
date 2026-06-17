package shared_errors

import "errors"

var (
	ErrInvalidAccessToken           = errors.New("invalid access token")
	ErrInvalidSessionToken          = errors.New("invalid session token")
	ErrFailedAssertionContextValues = errors.New("failed to assert type ContextValues: ")
	ErrCriticalServer               = errors.New("critical error on the server side")
	ErrIncorrectTypeRemove          = errors.New("the type  must be a soft-delete ot hard-delete")
)
