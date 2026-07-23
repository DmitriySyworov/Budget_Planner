package validate_password

import (
	"app/auth-service/internal/custom_errors"
	"strings"

	"github.com/nbutton23/zxcvbn-go"
)

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
