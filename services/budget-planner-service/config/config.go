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
	ApiPort   string
	Signature string
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
	signature := os.Getenv("JWT_SIGNATURE")
	counterEmptyVariables := 0
	if apiPort == "" {
		apiPort = "8080"
		logger.Warn("environment variable 'EXTERNAL_API_PORT' not found. Default value = 8080")
	}
	if signature == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'JWT_SIGNATURE' not found")
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
			ApiPort:   apiPort,
			Signature: signature,
		},
		DB: &DB{
			DSN: dsn,
		},
	}
}
