package shared_jwt

import (
	"errors"
	"shared/shared_errors"

	"github.com/golang-jwt/jwt/v4"
)

type SharedJWT struct {
	Signature []byte
}

func NewSharedJWT(signature string) *SharedJWT {
	return &SharedJWT{
		Signature: []byte(signature),
	}
}

var (
	ErrExpiredAccessToken = errors.New("access token has expired")
)

func (j *SharedJWT) ParseAccessToken(accessToken string) (string, error) {
	token, errToken := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		return j.Signature, nil
	})
	if errToken != nil {
		if errors.Is(errToken, jwt.ErrTokenExpired) {
			return "", ErrExpiredAccessToken
		}
		return "", shared_errors.ErrInvalidAccessToken
	}
	if types, okType := token.Claims.(jwt.MapClaims)["type"].(string); !okType || types != "access" {
		return "", shared_errors.ErrInvalidAccessToken
	}
	if userUUID, okUUID := token.Claims.(jwt.MapClaims)["sub"].(string); !okUUID {
		return "", shared_errors.ErrInvalidAccessToken
	} else {
		return userUUID, nil
	}
}
