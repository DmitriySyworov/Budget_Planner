package config

import (
	"app/budget-planner/internal/loggers"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	*Api
	*DB
}
type Api struct {
	ApiPort string
}
type DB struct {
	DSN           string
	RedisAddress  string
	RedisPassword string
}

func NewConfig(logger *loggers.Logger) *Config {
	errEnvFile := godotenv.Load()
	if errEnvFile != nil {
		logger.Warn(".env file not found.  This is normal if running inside a container")
	}
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	dsn := os.Getenv("DSN")
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	counterEmptyVariables := 0
	//if apiPort == "" {
	//	counterEmptyVariables++
	//	logger.Error("environment variable 'EXTERNAL_API_PORT' not found")
	//}
	//if dsn == "" {
	//	counterEmptyVariables++
	//	logger.Error("environment variable 'DSN' not found")
	//}
	//if redisAddress == "" {
	//	counterEmptyVariables++
	//	logger.Error("environment variable 'REDIS_ADDRESS' not found")
	//}
	//if redisPassword == "" {
	//	counterEmptyVariables++
	//	logger.Error("environment variable 'REDIS_PASSWORD' not found")
	//}
	if counterEmptyVariables != 0 {
		os.Exit(1)
	}
	return &Config{
		Api: &Api{
			ApiPort: apiPort,
		},
		DB: &DB{
			DSN:           dsn,
			RedisAddress:  redisAddress,
			RedisPassword: redisPassword,
		},
	}
}
