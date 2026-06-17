package auth

import "errors"

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrIncorrectAction  = errors.New("the action must be register, login or recovery")
	ErrCreateUser       = errors.New("failed to create user")
	ErrSentRefresh      = errors.New("refresh token not sent")
	ErrRenewalRefresh   = errors.New("refresh token renewal error")
	ErrIncorrectName    = errors.New("name must be between 2 and 64 characters")
)
