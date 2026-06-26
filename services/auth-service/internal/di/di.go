package di

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/model"
)

type IRepoUser interface {
	CreateUser(user *model.Users) error
	UserExistsByEmail(email string) bool
	GetPasswordByEmail(email string) (string, error)
	GetUserUUIDByEmail(email string) (string, error)
	RecoveryUser(userUUID string) error
	UpdateUser(user *model.Users, userUUID string) error
	GetUserByEmail(email string) (*model.Users, error)
}
type IServiceAuth interface {
	HelperAuth(action string, dataUser map[string]string) (*common.ResponseAuth, error)
}
type IRepoAuth interface {
	GetUserSession(sessionID, action string) (map[string]string, error)
}
