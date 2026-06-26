package user

import (
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/model"
	"shared/loggers"
	"shared/open_db"

	"gorm.io/gorm/clause"
)

type RepositoryUser struct {
	*open_db.Postgres
	*loggers.Logger
}

func NewRepositoryUser(postgres *open_db.Postgres, logger *loggers.Logger) *RepositoryUser {
	return &RepositoryUser{
		Postgres: postgres,
		Logger:   logger,
	}
}
func (r *RepositoryUser) CreateUser(user *model.Users) error {
	if errCreate := r.Postgres.Create(user).Error; errCreate != nil {
		r.Logger.Error("failed to create user: ", errCreate)
		return errCreate
	}
	return nil
}
func (r *RepositoryUser) UpdateUser(user *model.Users, userUUID string) error {
	if errUpdate := r.Postgres.
		Clauses(clause.Returning{}).
		Where("user_uuid = ?", userUUID).
		Updates(user).Error; errUpdate != nil {
		r.Logger.Error("failed to update user: ", errUpdate)
		return errUpdate
	}
	return nil
}
func (r *RepositoryUser) UserExistsByEmail(email string) bool {
	var isExist bool
	if errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM users
				 WHERE email = ?)`, email).Scan(&isExist).Error; errQuery != nil {
		r.Logger.Error("failed to check if the user exists by email: ", errQuery)
		return false
	}
	if !isExist {
		return false
	}
	return true
}
func (r *RepositoryUser) UserExistsByUserUUID(userUUID string) bool {
	var isExist bool
	if errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM users
				 WHERE user_uuid = ?)`, userUUID).Scan(&isExist).Error; errQuery != nil {
		r.Logger.Error("failed to check if the user exists by user_uuid: ", errQuery)
		return false
	}
	if !isExist {
		return false
	}
	return true
}
func (r *RepositoryUser) GetResponseUserByUUID(userUUID string) (*ResponseUser, error) {
	user := &ResponseUser{}
	if errGet := r.Postgres.Raw(`SELECT created_at, updated_at, name, email, user_uuid FROM users
WHERE user_uuid = ? AND deleted_at IS NULL`, userUUID).Scan(user).Error; errGet != nil {
		r.Logger.Error("failed to get user: " + errGet.Error())
		return nil, errGet
	}
	if user.UserUUID == "" {
		return nil, custom_errors.ErrNotFoundUser
	}
	return user, nil
}
func (r *RepositoryUser) GetUserByUUID(userUUID string) (*model.Users, error) {
	user := &model.Users{}
	if errGet := r.Postgres.Where("user_uuid = ?", userUUID).Take(user).Error; errGet != nil {
		return nil, errGet
	}
	return user, nil
}
func (r *RepositoryUser) GetUserByEmail(email string) (*model.Users, error) {
	user := &model.Users{}
	if errGet := r.Postgres.Where("email = ?", email).Take(user).Error; errGet != nil {
		return nil, errGet
	}
	return user, nil
}
func (r *RepositoryUser) GetPasswordByEmail(email string) (string, error) {
	var password string
	if errGetPassword := r.Postgres.Raw(`SELECT password FROM users
						WHERE email = ?`, email).Scan(&password).Error; errGetPassword != nil {
		r.Logger.Error("failed to get user password: ", errGetPassword)
		return "", ErrFailedGetUser
	}
	if password == "" {
		return "", custom_errors.ErrNotFoundUser
	}
	return password, nil
}
func (r *RepositoryUser) GetUserUUIDByEmail(email string) (string, error) {
	var userUUID string
	if errGetUserUUID := r.Postgres.Raw(`SELECT user_uuid FROM users
						WHERE email = ?`, email).Scan(&userUUID).Error; errGetUserUUID != nil {
		r.Logger.Error("failed to get userUUID: " + errGetUserUUID.Error())
		return "", ErrFailedGetUser
	}
	if userUUID == "" {
		return "", custom_errors.ErrNotFoundUser
	}
	return userUUID, nil
}
func (r *RepositoryUser) RemoveUser(userUUID string) error {
	if errRemove := r.Postgres.Where("user_uuid = ?", userUUID).Delete(&model.Users{}).Error; errRemove != nil {
		r.Logger.Error("failed to remove user: ", errRemove)
		return errRemove
	}
	return nil
}
func (r *RepositoryUser) DeleteUser(userUUID string) error {
	if errDelete := r.Postgres.
		Unscoped().
		Where("user_uuid = ?", userUUID).
		Delete(&model.Users{}).Error; errDelete != nil {
		r.Logger.Error("failed to delete user: ", errDelete)
		return errDelete
	}
	return nil
}
func (r *RepositoryUser) RecoveryUser(userUUID string) error {
	if errRecovery := r.Postgres.
		Model(&model.Users{}).
		Unscoped().
		Where("user_uuid = ?", userUUID).
		Update("deleted_at", nil).Error; errRecovery != nil {
		r.Logger.Error("failed to recovery user: " + errRecovery.Error())
		return errRecovery
	}
	return nil
}
