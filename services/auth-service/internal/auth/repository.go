package auth

import (
	"app/auth-service/internal/common"
	"context"
	"fmt"
	"shared/loggers"
	"shared/open_db"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RepositoryAuth struct {
	*open_db.Redis
	*loggers.Logger
}

const (
	session         = "session:"
	dataUserSession = "data_user:"
	refresh         = "refresh:"
)

func NewRepository(redis *open_db.Redis, logger *loggers.Logger) *RepositoryAuth {
	return &RepositoryAuth{
		Redis:  redis,
		Logger: logger,
	}
}
func (r *RepositoryAuth) CreateSession(sessionID string, code int) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	if errSet := r.Redis.Set(ctxTimeout, session+sessionID, code, time.Minute*5).Err(); errSet != nil {
		r.Logger.Error("failed to create session: ", errSet)
		return errSet
	}
	return nil
}
func (r *RepositoryAuth) CreateDataUserSession(name, email, hashPassword, sessionID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := dataUserSession + sessionID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSet := pipeliner.HSet(ctxTimeout, key, "name", name, "email", email, "password", hashPassword).
			Err(); errHSet != nil {
			r.Logger.Error("failed to create data user session: ", errHSet)
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

type DataUserSession struct {
	Name     string
	Email    string
	Password string
}

func (r *RepositoryAuth) GetDataUserSession(sessionID string) (*DataUserSession, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := dataUserSession + sessionID
	sessionValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, key).Result()
	if errHGetAll != nil {
		return nil, errHGetAll
	}
	return &DataUserSession{
		Name:     sessionValue["name"],
		Email:    sessionValue["email"],
		Password: sessionValue["password"],
	}, nil
}
func (r *RepositoryAuth) GetSession(sessionID string) (int, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	codeStr, errGetSession := r.Redis.Get(ctxTimeout, session+sessionID).Result()
	if errGetSession != nil {
		return 0, errGetSession
	}
	code, errParse := strconv.Atoi(codeStr)
	if errParse != nil {
		r.Logger.Error("failed to parse code: ", errParse)
		return 0, errParse
	}
	return code, nil
}
func (r *RepositoryAuth) CreateRefresh(refreshID, userUUID,  userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := refresh+refreshID
	r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSetRefresh := r.Redis.HSet(ctxTimeout, key, "userAgent", userAgent, "userUUID", userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh token: ", errHSetRefresh)
			return errHSetRefresh
		}
		if errExpired := r.Redis.Expire(ctxTimeout, key, common.TimeMonth)
		return nil
	})
	return nil
}
func (r *RepositoryAuth) GetRefresh(refreshID string) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	userAgent, errGet := r.Redis.Get(ctxTimeout, refresh+refreshID).Result()
	if errGet != nil {
		return "", errGet
	}
	return userAgent, nil
}
