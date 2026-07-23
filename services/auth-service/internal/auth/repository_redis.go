package auth

import (
	"app/auth-service/internal/common"
	"context"
	"shared/loggers"
	"shared/open_db"
	"shared/shared_constant"

	"github.com/redis/go-redis/v9"
)

type IRepositoryRedis interface {
	CreateUserSession(sessionID, action string, dataUser map[string]string) error
	CreateRefresh(userUUID string, refreshID, userAgent string) error
	DeleteRefresh(userUUID string, refreshID string) error
	GetRefresh(refreshID string) (*RefreshTokenValues, error)
	GetRefreshID(userUUID string) (string, error)
	RotationRefresh(userUUID, newRefreshID, OldRefreshID, userAgent string) error
	GetUserSession(sessionID, action string) (map[string]string, error)
}
type RepositoryRedisAuth struct {
	*open_db.Redis
	*loggers.Logger
}

func NewRepositoryRedis(redis *open_db.Redis, logger *loggers.Logger) IRepositoryRedis {
	return &RepositoryRedisAuth{
		Redis:  redis,
		Logger: logger,
	}
}

const (
	sessionKey = "session:"

	refreshKey     = "refresh:"
	userRefreshKey = "user_refresh:"
	userUUIDKey    = "user_uuid"
	userAgentKey   = "user_agent"
)

func (r *RepositoryRedisAuth) CreateUserSession(sessionID, action string, dataUser map[string]string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	keySession := sessionKey + action + ":" + sessionID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errHSet := pipeliner.HSet(ctxTimeout, keySession, dataUser).Err(); errHSet != nil {
			return errHSet
		}
		if errExpire := pipeliner.Expire(ctxTimeout, keySession, common.TTLSessionJWT).Err(); errExpire != nil {
			return errExpire
		}
		return nil
	}); errTx != nil {
		r.Logger.Error("failed to create user session: " + errTx.Error())
		return errTx
	}
	return nil
}
func (r *RepositoryRedisAuth) GetUserSession(sessionID, action string) (map[string]string, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	keySession := sessionKey + action + ":" + sessionID
	pipeline := r.Redis.Pipeline()
	mapCmd := pipeline.HGetAll(ctxTimeout, keySession)
	pipeline.Del(ctxTimeout, keySession)
	_, errPipeline := pipeline.Exec(ctxTimeout)
	if errPipeline != nil {
		r.Logger.Error("failed to get user session: " + errPipeline.Error())
		return nil, errPipeline
	}
	dataUser, errGetValues := mapCmd.Result()
	if errGetValues != nil {
		r.Logger.Error("failed to get user session: " + errGetValues.Error())
		return nil, errGetValues
	}
	return dataUser, nil
}

func (r *RepositoryRedisAuth) CreateRefresh(userUUID string, refreshID, userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	key := refreshKey + refreshID
	keyUserRefresh := userRefreshKey + userUUID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errSet := pipeliner.Set(ctxTimeout, keyUserRefresh, refreshID, common.TTLRefreshKey).Err(); errSet != nil {
			r.Logger.Error("failed to create user refresh: " + errSet.Error())
			return errSet
		}
		if errHSetRefresh := pipeliner.HSet(ctxTimeout, key, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
			r.Logger.Error("failed to create refresh session: " + errHSetRefresh.Error())
			return errHSetRefresh
		}
		if errExpire := pipeliner.Expire(ctxTimeout, key, common.TTLRefreshKey).Err(); errExpire != nil {
			r.Logger.Error("failed to add expiration time to key " + key + ": " + errExpire.Error())
			return errExpire
		}
		return nil
	}); errTx != nil {
		return errTx
	}
	return nil
}

func (r *RepositoryRedisAuth) DeleteRefresh(userUUID, refreshID string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	keyUserRefresh := userRefreshKey + userUUID
	keyRefresh := refreshKey + refreshID
	if _, errTx := r.Redis.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
		if errDel := pipeliner.Del(ctxTimeout, keyRefresh).Err(); errDel != nil {
			r.Logger.Error("failed to delete refreshID: " + errDel.Error())
			return errDel
		}
		if errDel := pipeliner.Del(ctxTimeout, keyUserRefresh).Err(); errDel != nil {
			r.Logger.Error("failed to delete user_refresh: " + errDel.Error())
			return errDel
		}
		return nil
	}); errTx != nil {
		return errTx
	}
	return nil
}

type RefreshTokenValues struct {
	UserUUID  string
	UserAgent string
}

func (r *RepositoryRedisAuth) GetRefresh(refreshID string) (*RefreshTokenValues, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	keyRefresh := refreshKey + refreshID
	refreshValue, errHGetAll := r.Redis.HGetAll(ctxTimeout, keyRefresh).Result()
	if errHGetAll != nil {
		r.Logger.Error("failed to get refresh session: " + errHGetAll.Error())
		return nil, errHGetAll
	}
	if len(refreshValue) == 0 {
		r.Logger.Error("not found refresh session")
		return nil, ErrRenewalRefresh
	}
	return &RefreshTokenValues{
		UserAgent: refreshValue[userAgentKey],
		UserUUID:  refreshValue[userUUIDKey],
	}, nil
}
func (r *RepositoryRedisAuth) GetRefreshID(userUUID string) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	keyUserRefresh := userRefreshKey + userUUID
	refreshID, errGet := r.Redis.Get(ctxTimeout, keyUserRefresh).Result()
	if errGet != nil {
		r.Logger.Warn("failed to get refreshID: " + errGet.Error())
		return "", errGet
	}
	if refreshID == "" {
		r.Logger.Warn("not found refresh session")
		return "", ErrRenewalRefresh
	}
	return refreshID, nil
}
func (r *RepositoryRedisAuth) RotationRefresh(userUUID, newRefreshID, OldRefreshID, userAgent string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_constant.CtxTimeoutRedis)
	defer cancel()
	newRefreshKey := refreshKey + newRefreshID
	oldRefreshKey := refreshKey + OldRefreshID
	keyUserRefresh := userRefreshKey + userUUID

	if errWatch := r.Redis.Watch(ctxTimeout, func(tx *redis.Tx) error {
		exist, errCheckExistOLdRefreshID := tx.Exists(ctxTimeout, oldRefreshKey).Result()
		if errCheckExistOLdRefreshID != nil {
			r.Logger.Warn("attempted to refresh an expired or non-existent token")
			return errCheckExistOLdRefreshID
		}
		if exist == 0 {
			return redis.Nil
		}
		if _, erTx := tx.TxPipelined(ctxTimeout, func(pipeliner redis.Pipeliner) error {
			if errDel := pipeliner.Del(ctxTimeout, oldRefreshKey).Err(); errDel != nil {
				r.Logger.Error("failed to delete refresh session: " + errDel.Error())
				return errDel
			}
			if errDel := pipeliner.Del(ctxTimeout, keyUserRefresh).Err(); errDel != nil {
				r.Logger.Error("failed to delete user_refresh: " + errDel.Error())
				return errDel
			}
			if errSet := pipeliner.Set(ctxTimeout, keyUserRefresh, newRefreshID, common.TTLRefreshKey).Err(); errSet != nil {
				r.Logger.Error("failed to create user refresh: " + errSet.Error())
				return errSet
			}
			if errHSetRefresh := pipeliner.HSet(ctxTimeout, newRefreshKey, userAgentKey, userAgent, userUUIDKey, userUUID).Err(); errHSetRefresh != nil {
				r.Logger.Error("failed to create refresh session: " + errHSetRefresh.Error())
				return errHSetRefresh
			}
			if errExpire := pipeliner.Expire(ctxTimeout, newRefreshKey, common.TTLRefreshKey).Err(); errExpire != nil {
				r.Logger.Error("failed to add expiration time to key " + newRefreshKey + ": " + errExpire.Error())
				return errExpire
			}
			return nil
		}); erTx != nil {
			return erTx
		}
		return nil
	}, oldRefreshKey); errWatch != nil {
		return ErrRefreshExpired
	}
	return nil
}
