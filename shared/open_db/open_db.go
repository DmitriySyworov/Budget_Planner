package open_db

import (
	"os"
	"shared/loggers"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Redis struct {
	*redis.Client
}
type Postgres struct {
	*gorm.DB
}

func OpenPostgres(DSN string, loggers *loggers.Logger) *Postgres {
	db, errOpen := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if errOpen != nil {
		loggers.Error("failed to connect PostgreSQL: " + errOpen.Error())
		os.Exit(1)
	}
	sqlDriver, errExtractSQL := db.DB()
	if errExtractSQL != nil {
		loggers.Error("failed to extract driver sql: " + errExtractSQL.Error())
		os.Exit(1)
	}
	sqlDriver.SetMaxOpenConns(100)
	sqlDriver.SetMaxIdleConns(20)
	sqlDriver.SetConnMaxLifetime(1 * time.Hour)
	sqlDriver.SetConnMaxIdleTime(10 * time.Minute)
	loggers.Info("connect MySQL successful")
	return &Postgres{
		DB: db,
	}
}
func OpenRedis(redisAddress, redisPassword string, loggers *loggers.Logger) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:            redisAddress,
		Password:        redisPassword,
		DB:              0,
		PoolSize:        100,
		MinIdleConns:    20,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	})
	loggers.Info("connect redis successful")
	return &Redis{
		Client: rdb,
	}
}
