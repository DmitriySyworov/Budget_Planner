package auth

import "errors"

var (
	ErrFailedSecurity           = errors.New("failed to ensure security")
	ErrUserAlreadyExist         = errors.New("user already exist")
	ErrIncorrectPasswordOrEmail = errors.New("incorrect password or email")
	ErrIncorrectSessionID       = errors.New("incorrect session id")
	ErrIncorrectAction          = errors.New("the action must be register, login or recovery")
	ErrSessionExpired           = errors.New("authorization session has expired")
	ErrCreateUser               = errors.New("failed to create user")
	ErrSentRefresh              = errors.New("refresh token not sent")
	ErrRenewalRefresh           = errors.New("refresh token renewal error")
	ErrIncorrectFormatCode      = errors.New("the code must be 6 characters")
	ErrIncorrectCode            = errors.New("incorrect code")
	ErrIncorrectName            = errors.New("name must be between 2 and 64 characters")
	ErrIncorrectEnterPassword   = errors.New("password must be between 8 and 24 characters")
	ErrIncorrectEmail           = errors.New("incorrect email")
)
