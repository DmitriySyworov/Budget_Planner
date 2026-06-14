package custom_errors

import "errors"

var (
	ErrNotFoundUser = errors.New("not found user")
)
