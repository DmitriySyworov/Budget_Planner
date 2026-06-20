package open_db

import (
	"os"
	"shared/loggers"

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
		loggers.Error("failed to connect PostgreSQL: ", errOpen)
		os.Exit(1)
	}
	return &Postgres{
		DB: db,
	}
}
func OpenRedis(redisAddress, redisPassword string) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
	})
	return &Redis{
		Client: rdb,
	}
}
