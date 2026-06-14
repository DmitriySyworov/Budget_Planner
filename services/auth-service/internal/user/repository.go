package user

import (
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/model"
	"shared/loggers"
	"shared/open_db"
)

type RepositoryUser struct {
	*open_db.Postgres
	*loggers.Logger
}

func NewRepositoryUser(postgres *open_db.Postgres) *RepositoryUser {
	return &RepositoryUser{
		Postgres: postgres,
	}
}
func (r *RepositoryUser) CreateUser(user *model.User) error {
	if errCreate := r.Postgres.Create(user).Error; errCreate != nil {
		r.Logger.Error("failed to create user: ", errCreate)
		return errCreate
	}
	return nil
}
func (r *RepositoryUser) UserExistsByEmail(email string) bool {
	var isExist bool
	errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM users
				 WHERE email = ?)`, email).Scan(&isExist)
	if !isExist || errQuery != nil {
		return false
	}
	return true
}

func (r *RepositoryUser) UserExistsByUUID(userUUID string) bool {
	var isExist bool
	errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM users
				 WHERE user_uuid = ?)`, userUUID).Scan(&isExist)
	if !isExist || errQuery != nil {
		return false
	}
	return true
}
func (r *RepositoryUser) GetPasswordByEmail(email string) (string, error) {
	var password string
	if errGetPassword := r.Postgres.Raw(`SELECT password FROM users
						WHERE email = ?`, email).Scan(&password).Error; errGetPassword != nil {
		r.Logger.Error("failed to get password: ", errGetPassword)
		return "", errGetPassword
	}
	if password == "" {
		return "", custom_errors.ErrNotFoundUser
	}
	return password, nil
}
