package common

import (
	"app/auth-service/internal/custom_errors"
	"strings"
	"time"

	"github.com/nbutton23/zxcvbn-go"
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

func ValidatePassword(password, email string, userInputs []string) error {
	if strings.Contains(password, email) {
		return custom_errors.ErrPasswordContainEmail
	}
	res := zxcvbn.PasswordStrength(password, userInputs)
	if res.Score < 3 {
		return custom_errors.ErrPasswordIsNotStrong
	}
	return nil
}
