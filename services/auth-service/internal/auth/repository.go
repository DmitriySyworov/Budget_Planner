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
	sessionKey      = "session:"
	dataUserAuthKey = "data_user_auth:"
	refreshKey      = "refresh:"
	nameKey         = "name"
	emailKey        = "email"
	passwordKey     = "password"
	userUUIDKey     = "user_uuid"
	userAgentKey    = "user_agent"
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
	if errSet := r.Redis.Set(ctxTimeout, sessionKey+sessionID, code, time.Minute*5).Err(); errSet != nil {
		r.Logger.Error("failed to create session: ", errSet)
		return errSet
	}
	return nil
}
func (r *RepositoryAuth) CreateDataUserSession(name, email, hashPassword, sessionID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := dataUserAuthKey + sessionID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSet := pipeliner.HSet(ctxTimeout, key, nameKey, name, emailKey, email, passwordKey, hashPassword).
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
	key := dataUserAuthKey + sessionID
	sessionValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, key).Result()
	if errHGetAll != nil {
		return nil, errHGetAll
	}
	return &DataUserSession{
		Name:     sessionValue[nameKey],
		Email:    sessionValue[emailKey],
		Password: sessionValue[passwordKey],
	}, nil
}
func (r *RepositoryAuth) GetSession(sessionID string) (int, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	codeStr, errGetSession := r.Redis.Get(ctxTimeout, sessionKey+sessionID).Result()
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
func (r *RepositoryAuth) CreateRefresh(refreshID, userUUID, userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := refreshKey + refreshID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSetRefresh := r.Redis.HSet(ctxTimeout, key, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh token: ", errHSetRefresh)
			return errHSetRefresh
		}
		if errExpire := r.Redis.Expire(ctxTimeout, key, common.TimeMonth).Err(); errExpire != nil {
			r.Logger.Error(fmt.Sprintf("failed to add expiration time to key %s: ", key), errExpire)
			return errExpire
		}
		return nil
	}); errTx != nil {
		return errTx
	}
	return nil
}

type DtoRefreshToken struct {
	UserUUID  string
	UserAgent string
}

func (r *RepositoryAuth) GetRefresh(refreshID string) (*DtoRefreshToken, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	refreshValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, refreshKey+refreshID).Result()
	if errHGetAll != nil {
		return nil, errHGetAll
	}
	return &DtoRefreshToken{
		UserAgent: refreshValue[userAgentKey],
		UserUUID:  refreshValue[userUUIDKey],
	}, nil
}
func (r *RepositoryAuth) RotationRefresh(newRefreshID, OldRefreshID, userUUID, userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	newRefreshKey := refreshKey + newRefreshID
	if _, erTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errDel := pipeliner.Del(ctxTimeout, refreshKey+OldRefreshID).Err(); errDel != nil {
			r.Logger.Error("failed to delete refresh token: ", errDel)
			return errDel
		}
		if errHSetRefresh := r.Redis.HSet(ctxTimeout, newRefreshKey, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh token: ", errHSetRefresh)
			return errHSetRefresh
		}
		if errExpire := r.Redis.Expire(ctxTimeout, newRefreshKey, common.TimeMonth).Err(); errExpire != nil {
			r.Logger.Error(fmt.Sprintf("failed to add expiration time to key %s: ", newRefreshKey), errExpire)
			return errExpire
		}
		return nil
	}); erTx != nil {
		return erTx
	}
	return nil
}
