package JWT

import (
	"app/auth-service/internal/common"
	"shared/loggers"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	Logger    *loggers.Logger
	Signature []byte
}

func NewJWT(signature string, logger *loggers.Logger) *JWT {
	return &JWT{
		Signature: []byte(signature),
		Logger:    logger,
	}
}
func (j *JWT) CreateSessionJWT(sessionID string) (string, error) {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_id": sessionID,
		"expires_at": time.Now().Add(time.Minute * 5),
	})
	token, errToken := claim.SignedString(j.Signature)
	if errToken != nil {
		j.Logger.Error("failed to sign token: ", errToken)
		return "", nil
	}
	return token, nil
}
func (j *JWT) CreateAccessJWT(accessID, userUUID string) (string, error) {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type":       "access",
		"access_id":  accessID,
		"user_uuid":  userUUID,
		"expires_at": time.Now().Add(time.Minute * 5),
	})
	token, errToken := claim.SignedString(j.Signature)
	if errToken != nil {
		j.Logger.Error("failed to sign token: ", errToken)
		return "", nil
	}
	return token, nil
}
func (j *JWT) CreateRefreshJWT(refreshID string) (string, error) {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type":       "refresh",
		"refresh_id": refreshID,
		"expires_at": time.Now().Add(common.TimeMonth),
	})
	token, errToken := claim.SignedString(j.Signature)
	if errToken != nil {
		j.Logger.Error("failed to sign token: ", errToken)
		return "", nil
	}
	return token, nil
}
