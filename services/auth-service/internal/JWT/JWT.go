package JWT

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"errors"
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
		"type": "session",
		"sub":  sessionID,
		"exp":  time.Now().Add(time.Minute * 5).Unix(),
	})
	token, errToken := claim.SignedString(j.Signature)
	if errToken != nil {
		j.Logger.Error("failed to sign token: ", errToken)
		return "", nil
	}
	return token, nil
}
func (j *JWT) CreateAccessJWT(userUUID string) (string, error) {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type": "access",
		"sub":  userUUID,
		"exp":  time.Now().Add(time.Minute * 5).Unix(),
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
		"type": "refresh",
		"sub":  refreshID,
		"exp":  time.Now().Add(common.TimeMonth).Unix(),
	})
	token, errToken := claim.SignedString(j.Signature)
	if errToken != nil {
		j.Logger.Error("failed to sign token: ", errToken)
		return "", nil
	}
	return token, nil
}

var (
	ErrExpiredRefreshToken   = errors.New("refresh token has expired")
	ErrIncorrectRefreshToken = errors.New("incorrect refresh token")
)

func (j *JWT) ParseRefreshToken(refreshToken string) (string, error) {
	token, errToken := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		return j.Signature, nil
	})
	if errToken != nil {
		if errors.Is(errToken, jwt.ErrTokenExpired) {
			return "", ErrExpiredRefreshToken
		}
		return "", ErrIncorrectRefreshToken
	}
	if types, okType := token.Claims.(jwt.MapClaims)["type"].(string); !okType || types != "refresh" {
		return "", ErrIncorrectRefreshToken
	}
	if userUUID, okUUID := token.Claims.(jwt.MapClaims)["sub"].(string); !okUUID {
		return "", ErrIncorrectRefreshToken
	} else {
		return userUUID, nil
	}
}
func (j *JWT) ParseSessionToken(accessToken string) (string, error) {
	token, errToken := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return j.Signature, nil
	})
	if errToken != nil {
		return "", custom_errors.ErrInvalidSessionToken
	}
	if types, okType := token.Claims.(jwt.MapClaims)["type"].(string); !okType || types != "session" {
		return "", custom_errors.ErrInvalidSessionToken
	}
	if sessionID, okSessionID := token.Claims.(jwt.MapClaims)["sub"].(string); !okSessionID {
		return "", custom_errors.ErrInvalidSessionToken
	} else {
		return sessionID, nil
	}
}
