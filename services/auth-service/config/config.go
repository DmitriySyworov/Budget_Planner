package authconfig

import (
	"os"
	"shared/loggers"

	"github.com/joho/godotenv"
)

type Config struct {
	*API
	*Db
	*VerifyEmail
}

type API struct {
	ApiPort string
}
type VerifyEmail struct {
	ApiEmail        string
	ApiPassword     string
	SmtpAddress     string
	SmtpAddressHost string
	Signature       string
}
type Db struct {
	DSN           string
	RedisAddress  string
	RedisPassword string
}

func NewConfig(logger *loggers.Logger) *Config {
	if errFileEnv := godotenv.Load(); errFileEnv != nil {
		logger.Warn(".env file not found.  This is normal if running inside a container")
	}
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	signature := os.Getenv("JWT_SIGNATURE")
	dsn := os.Getenv("DSN")
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	apiEmail := os.Getenv("API_EMAIL")
	apiPassword := os.Getenv("API_PASSWORD")
	smtpAddress := os.Getenv("SMTP_ADDRESS")
	smtpAddressHost := os.Getenv("SMTP_ADDRESS_HOST")
	var counterEmptyVariables int
	if apiPort == "" {
		apiPort = "8080"
		logger.Warn("environment variable 'EXTERNAL_API_PORT' not found. Default value = 8080")
	}
	if dsn == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'DSN' not found")
	}
	if redisAddress == "" {
		redisAddress = "localhost:6379"
		logger.Error("environment variable 'REDIS_ADDRESS' not found. Default value = localhost:6379")
	}
	if redisPassword == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'REDIS_PASSWORD' not found")
	}
	if signature == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'JWT_SIGNATURE' not found")
	}
	if apiEmail == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'API_EMAIL' not found")
	}
	if apiPassword == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'API_PASSWORD' not found")
	}
	if smtpAddress == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'SMTP_ADDRESS' not found")
	}
	if smtpAddressHost == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'SMTP_ADDRESS_HOST' not found")
	}
	if counterEmptyVariables != 0 {
		os.Exit(1)
	}
	return &Config{
		Db: &Db{
			DSN:           dsn,
			RedisPassword: redisPassword,
			RedisAddress:  redisAddress,
		},
		API: &API{
			ApiPort: apiPort,
		},
		VerifyEmail: &VerifyEmail{
			ApiEmail:        apiEmail,
			ApiPassword:     apiPassword,
			SmtpAddress:     smtpAddress,
			SmtpAddressHost: smtpAddressHost,
			Signature:       signature,
		},
	}
}
