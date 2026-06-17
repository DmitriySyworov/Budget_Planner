package di

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/common"
	"app/auth-service/internal/model"
)

type IRepoUser interface {
	CreateUser(user *model.User) error
	UserExistsByEmail(email string) bool
	GetPasswordByEmail(email string) (string, error)
}
type IServiceAuth interface {
	HelperAuth(userEmail string, conf *authconfig.VerifyEmail) (*common.ResponseAuth, string, error)
}
type IRepoAuth interface {
	GetSession(sessionID string) (int, error)
}
