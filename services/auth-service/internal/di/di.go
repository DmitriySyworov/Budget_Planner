package di

import "app/auth-service/internal/model"

type IRepoUser interface {
	CreateUser(user *model.User) error
	UserExistsByEmail(email string) bool
	GetPasswordByEmail(email string) (string, error)
}
