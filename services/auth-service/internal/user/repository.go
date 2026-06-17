package user

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/model"
	"context"
	"fmt"
	"shared/loggers"
	"shared/open_db"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm/clause"
)

type RepositoryUser struct {
	*open_db.Postgres
	*open_db.Redis
	*loggers.Logger
}

func NewRepositoryUser(postgres *open_db.Postgres, redis *open_db.Redis, logger *loggers.Logger) *RepositoryUser {
	return &RepositoryUser{
		Postgres: postgres,
		Logger:   logger,
		Redis:    redis,
	}
}
func (r *RepositoryUser) CreateUser(user *model.User) error {
	if errCreate := r.Postgres.Create(user).Error; errCreate != nil {
		r.Logger.Error("failed to create user: ", errCreate)
		return errCreate
	}
	return nil
}
func (r *RepositoryUser) UpdateUser(user *model.User, userUUID string) error {
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
	var user ResponseUser
	if errGet := r.Postgres.Raw(`SELECT created_at, updated_at, name, email, user_uuid FROM users
WHERE user_uuid = ?`, userUUID).Error; errGet != nil || user.UserUUID == "" {
		r.Logger.Error("failed to get user: ", errGet)
		return nil, errGet
	}
	if user.UserUUID == "" {
		return nil, custom_errors.ErrNotFoundUser
	}
	return &user, nil
}
func (r *RepositoryUser) GetUserByUUID(userUUID string) (*model.User, error) {
	var user model.User
	if errGet := r.Postgres.Where("user_uuid = ", userUUID).Take(&user).Error; errGet != nil {
		return nil, errGet
	}
	return &user, nil
}
func (r *RepositoryUser) GetPasswordByEmail(email string) (string, error) {
	var password string
	if errGetPassword := r.Postgres.Raw(`SELECT password FROM users
						WHERE email = ?`, email).Scan(&password).Error; errGetPassword != nil {
		r.Logger.Error("failed to get password: ", errGetPassword)
		r.Logger.Error("failed to get user password: ", errGetPassword)
		return "", ErrFailedGetUser
	}
	if password == "" {
		return "", custom_errors.ErrNotFoundUser
	}
	return password, nil
}
func (r *RepositoryUser) RemoveUser(userUUID string) error {
	if errRemove := r.Postgres.Where("user_uuid = ?", userUUID).Delete(&model.User{}).Error; errRemove != nil {
		r.Logger.Error("failed to remove user: ", errRemove)
		return errRemove
	}
	return nil
}
func (r *RepositoryUser) DeleteUser(userUUID string) error {
	if errDelete := r.Postgres.
		Unscoped().
		Where("user_uuid = ?", userUUID).
		Delete(&model.User{}).Error; errDelete != nil {
		r.Logger.Error("failed to delete user: ", errDelete)
		return errDelete
	}
	return nil
}

const (
	dataUpdateKey      = "data_user_update:"
	dataRemoveKey      = "data_user_remove:"
	newNameKey         = "new_name"
	newEmailKey        = "new_email"
	newHashPasswordKey = "new_hash_password"
)

func (r *RepositoryUser) CreateRemoveDataUserSession(typeRemove, sessionID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	if errSet := r.Redis.Set(ctxTimeout, dataRemoveKey+sessionID, typeRemove, time.Minute*5).Err(); errSet != nil {
		r.Logger.Error("failed to set session remove user: ", errSet)
		return errSet
	}
	return nil
}
func (r *RepositoryUser) GetRemoveDataUserSession(sessionID string) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	typeRemove, errGet := r.Redis.Get(ctxTimeout, dataRemoveKey+sessionID).Result()
	if errGet != nil {
		return "", errGet
	}
	return typeRemove, nil
}
func (r *RepositoryUser) CreateUpdateDataUserSession(newName, newEmail, newHashPassword, sessionID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := dataUpdateKey + sessionID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSet := pipeliner.HSet(ctxTimeout, key, newNameKey, newName, newEmailKey, newEmail, newHashPasswordKey, newHashPassword).
			Err(); errHSet != nil {
			r.Logger.Error("failed to create data update  user session: ", errHSet)
			return errHSet
		}
		if errExpire := pipeliner.Expire(ctxTimeout, key, time.Minute*5).Err(); errExpire != nil {
			r.Logger.Error(fmt.Sprintf("failed to add expiration time to key %s: ", key), errExpire)
			return errExpire
		}
		return nil
	}); errTx != nil {
		return errTx
	}
	return nil
}

type DataUserUpdateSession struct {
	NewName     string
	NewEmail    string
	NewPassword string
}

func (r *RepositoryUser) GetUpdateDataUserSession(sessionID string) (*DataUserUpdateSession, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := dataUpdateKey + sessionID
	sessionValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, key).Result()
	if errHGetAll != nil {
		return nil, errHGetAll
	}
	return &DataUserUpdateSession{
		NewName:     sessionValue[newNameKey],
		NewEmail:    sessionValue[newEmailKey],
		NewPassword: sessionValue[newHashPasswordKey],
	}, nil
}
