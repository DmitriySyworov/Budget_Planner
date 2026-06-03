package open_db

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/loggers"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OpenDB struct {
	*Redis
	*Postgres
}
type Redis struct {
	*redis.Client
}
type Postgres struct {
	*gorm.DB
}

func NewOpenDB(logger *loggers.Logger, conf *config.DB) *OpenDB {
	db, errOpen := gorm.Open(postgres.Open(conf.DSN))
	if errOpen != nil {
		logger.Error("failed to connect PostgreSQL")
		os.Exit(1)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddress,
		Password: conf.RedisPassword,
	})
	return &OpenDB{
		Redis: &Redis{
			Client: rdb,
		},
		Postgres: &Postgres{
			DB: db,
		},
	}
}
