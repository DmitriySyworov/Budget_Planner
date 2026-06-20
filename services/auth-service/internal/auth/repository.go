package auth

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"context"
	"errors"
	"fmt"
	"shared/loggers"
	"shared/open_db"
	"time"

	"github.com/redis/go-redis/v9"
)

type RepositoryAuth struct {
	*open_db.Redis
	*loggers.Logger
}

const (
	sessionKey     = "session_"
	refreshKey     = "refresh:"
	userRefreshKey = "user_refresh:"
	userUUIDKey    = "user_uuid"
	userAgentKey   = "user_agent"
)

func NewRepository(redis *open_db.Redis, logger *loggers.Logger) *RepositoryAuth {
	return &RepositoryAuth{
		Redis:  redis,
		Logger: logger,
	}
}

func (r *RepositoryAuth) CreateUserSession(sessionID, action string, dataUser map[string]string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := sessionKey + action + ":" + sessionID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSet := pipeliner.HSet(ctxTimeout, key, dataUser).
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

func (r *RepositoryAuth) GetUserSession(sessionID, action string) (map[string]string, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := sessionKey + action + ":" + sessionID
	sessionValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, key).Result()
	if errHGetAll != nil {
		r.Logger.Error("failed to get session: " + errHGetAll.Error())
		return nil, errHGetAll
	}
	if sessionValue[common.CodeKey] == "" {
		return nil, custom_errors.ErrSessionExpired
	}
	return sessionValue, nil
}
func (r *RepositoryAuth) CreateRefresh(refreshID, userUUID, userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	key := refreshKey + refreshID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errSet := r.Redis.Set(ctxTimeout, userRefreshKey+userUUID, refreshID, common.TimeMonth).Err(); errSet != nil {
			r.Logger.Error("failed to create user refresh: " + errSet.Error())
			return errSet
		}
		if errHSetRefresh := r.Redis.HSet(ctxTimeout, key, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh session: " + errHSetRefresh.Error())
			return errHSetRefresh
		}
		if errExpire := r.Redis.Expire(ctxTimeout, key, common.TimeMonth).Err(); errExpire != nil {
			r.Logger.Error(fmt.Sprintf("failed to add expiration time to key %s: ", key) + errExpire.Error())
			return errExpire
		}
		return nil
	}); errTx != nil {
		return errTx
	}
	return nil
}
func (r *RepositoryAuth) DeleteOldRefresh(userUUID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	refreshID, errGet := r.Redis.Get(ctxTimeout, userRefreshKey+userUUID).Result()
	if errGet != nil || refreshID == "" {
		r.Logger.Warn("failed to get user refresh: " + "not found user_refresh")
		return errors.New("not found user_refresh")
	}
	keyRefresh := refreshKey + refreshID
	if errDel := r.Redis.Del(ctxTimeout, keyRefresh).Err(); errDel != nil {
		r.Logger.Error("failed to delete old refreshID: " + errDel.Error())
		return errDel
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
		r.Logger.Error("failed to get refresh session: " + errHGetAll.Error())
		return nil, errHGetAll
	}
	if refreshValue[userUUIDKey] == "" {
		return nil, ErrRenewalRefresh
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
			r.Logger.Error("failed to delete refresh session: " + errDel.Error())
			return errDel
		}
		if errHSetRefresh := r.Redis.HSet(ctxTimeout, newRefreshKey, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh session: " + errHSetRefresh.Error())
			return errHSetRefresh
		}
		if errExpire := r.Redis.Expire(ctxTimeout, newRefreshKey, common.TimeMonth).Err(); errExpire != nil {
			r.Logger.Error(fmt.Sprintf("failed to add expiration time to key %s: ", newRefreshKey) + errExpire.Error())
			return errExpire
		}
		return nil
	}); erTx != nil {
		return erTx
	}
	return nil
}
