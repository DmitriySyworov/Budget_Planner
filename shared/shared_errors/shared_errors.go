package shared_errors

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAccessToken           = errors.New("invalid access token")
	ErrInvalidSessionToken          = errors.New("invalid session token")
	ErrFailedAssertionContextValues = errors.New("failed to assert type ContextValues: ")
	ErrCriticalServer               = errors.New("critical error on the server side")
	ErrIncorrectTypeRemove          = errors.New("the type  must be a soft-delete ot hard-delete")
	ErrIncorrectLimit               = errors.New("the limit must be a positive integer not greater than 100")
	ErrIncorrectOffset              = errors.New("the offset must be a positive integer")
)

type MapError struct {
	Map map[string]string
}

func (m MapError) Error() string {
	sliceError := make([]string, 0, 5)
	for _, value := range m.Map {
		sliceError = append(sliceError, value)
	}
	return strings.Join(sliceError, ";")
}
