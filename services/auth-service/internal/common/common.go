package common

import (
	"time"
)

const (
	CtxTimeout = time.Second * 3
	TimeMonth  = time.Hour * 720

	CodeKey = "code"
)

type ResponseAuth struct {
	Message    string
	JwtSession string `json:"jwt_session"`
}
type RequestConfirm struct {
	Code int `validate:"required,min=100000,max=999999"`
}
