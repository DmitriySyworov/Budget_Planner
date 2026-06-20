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
	GetUserUUIDByEmail(email string) (string, error)
	RecoveryUser(userUUID string) error
}
type IServiceAuth interface {
	HelperAuth(action string, dataUser map[string]string, conf *authconfig.VerifyEmail) (*common.ResponseAuth, error)
}
type IRepoAuth interface {
	GetUserSession(sessionID, action string) (map[string]string, error)
}
