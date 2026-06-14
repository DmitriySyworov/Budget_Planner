package auth

import "errors"

var (
	ErrFailedSecurity     = errors.New("failed to ensure security")
	ErrUserAlreadyExist   = errors.New("user already exist")
	ErrIncorrectPassword  = errors.New("incorrect password")
	ErrIncorrectCode      = errors.New("incorrect code")
	ErrIncorrectSessionID = errors.New("incorrect session id")
	ErrIncorrectAction    = errors.New("the action must be register, login or recovery")
	ErrSessionExpired     = errors.New("authorization session has expired")
	ErrCreateUser         = errors.New("failed to create user")
	ErrSentRefresh        = errors.New("refresh token not sent")
	ErrRenewalRefresh     = errors.New("refresh token renewal error")
)
