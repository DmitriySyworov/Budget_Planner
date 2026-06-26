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
	SessionJwt string `json:"session_jwt"`
}
