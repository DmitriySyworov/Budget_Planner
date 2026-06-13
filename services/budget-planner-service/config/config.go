package budgetconfig

import (
	"os"
	"shared/loggers"

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
	DSN string
}

func NewConfig(logger *loggers.Logger) *Config {
	errEnvFile := godotenv.Load()
	if errEnvFile != nil {
		logger.Warn(".env file not found.  This is normal if running inside a container")
	}
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	dsn := os.Getenv("DSN")
	counterEmptyVariables := 0
	if apiPort == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'EXTERNAL_API_PORT' not found")
	}
	if dsn == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'DSN' not found")
	}
	if counterEmptyVariables != 0 {
		os.Exit(1)
	}
	return &Config{
		Api: &Api{
			ApiPort: apiPort,
		},
		DB: &DB{
			DSN: dsn,
		},
	}
}
