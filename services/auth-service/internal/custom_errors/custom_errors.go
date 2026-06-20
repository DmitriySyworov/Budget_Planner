package custom_errors

import "errors"

var (
	ErrNotFoundUser             = errors.New("not found user")
	ErrIncorrectEmail           = errors.New("incorrect email")
	ErrFailedSecurity           = errors.New("failed to ensure security")
	ErrIncorrectPasswordOrEmail = errors.New("incorrect password or email")
	ErrIncorrectEnterPassword   = errors.New("password must be between 8 and 24 characters")
	ErrIncorrectFormatCode      = errors.New("the code must be 6 characters")
	ErrIncorrectSessionID       = errors.New("incorrect session id")
	ErrSessionExpired           = errors.New("authorization session has expired or does not exist")
	ErrIncorrectCode            = errors.New("incorrect code")
)
