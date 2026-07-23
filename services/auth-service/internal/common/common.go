package common

import (
	"time"
)

const (
	TimeMonth     = time.Hour * 720
	TTLRefreshKey = time.Hour * 720
	TTLSessionJWT = time.Minute * 5
	TTLAccessKey  = time.Minute * 15
	CodeKey       = "code"

	ServiceName = "budget-planner"
)

type ResponseAuth struct {
	Message    string
	SessionJwt string `json:"session_jwt"`
}
