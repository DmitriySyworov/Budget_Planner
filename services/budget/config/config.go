package budgetconfig

import (
	"os"
	"shared/loggers"

	"github.com/joho/godotenv"
)

type Config struct {
	*Api
	*DB
	*Kafka
	*SharedRedis
}
type Api struct {
	ApiPort   string
	ServiceIP string
	Signature string
}
type DB struct {
	DSN string
}
type SharedRedis struct {
	SharedRedisAddress  string
	SharedRedisPassword string
}
type Kafka struct {
	Broker              string
	KafkaUser           string
	KafkaPassword       string
	DeletedUsersTopic   string
	BudgetDeleteGroupID string
}

func NewConfig(logger *loggers.Logger) *Config {
	if godotenv.Load() != nil {
		logger.Warn(".env file not found.  This is normal if running inside a container")
		if godotenv.Load(".env.test") != nil {
			logger.Warn(".env.test file not found. This is normal if tests don't run")
		}
	}
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	serviceIP := os.Getenv("SERVICE_IP")
	dsn := os.Getenv("DSN")
	signature := os.Getenv("JWT_SIGNATURE")
	broker := os.Getenv("KAFKA_BOOTSTRAP")
	kafkaUser := os.Getenv("KAFKA_CLIENT_USER")
	kafkaPassword := os.Getenv("KAFKA_CLIENT_PASSWORD")
	deletedUsersTopic := os.Getenv("DELETED_USERS_TOPIC")
	budgetDeleteGroupID := os.Getenv("BUDGET_DELETE_GROUP_ID")
	sharedRedisAddress := os.Getenv("SHARED_REDIS_ADDRESS")
	sharedRedisPassword := os.Getenv("SHARED_REDIS_PASSWORD")
	counterEmptyVariables := 0
	if apiPort == "" {
		apiPort = "8080"
		logger.Warn("environment variable 'EXTERNAL_API_PORT' not found. Default value = 8080")
	}
	if serviceIP == "" {
		apiPort = "localhost"
		logger.Warn("environment variable 'SERVICE_IP' not found. Default value = localhost")
	}
	if signature == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'JWT_SIGNATURE' not found")
	}
	if dsn == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'DSN' not found")
	}
	if broker == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'KAFKA_BOOTSTRAP' not found")
	}
	if kafkaUser == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'KAFKA_CLIENT_USER' not found")
	}
	if kafkaPassword == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'KAFKA_CLIENT_PASSWORD' not found")
	}
	if deletedUsersTopic == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'DELETED_USERS_TOPIC' not found")
	}
	if budgetDeleteGroupID == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'BUDGET_DELETE_GROUP_ID' not found")
	}
	if sharedRedisAddress == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'SHARED_REDIS_ADDRESS' not found")
	}
	if sharedRedisPassword == "" {
		counterEmptyVariables++
		logger.Error("environment variable 'SHARED_REDIS_PASSWORD' not found")
	}
	if counterEmptyVariables != 0 {
		os.Exit(1)
	}
	return &Config{
		Api: &Api{
			ApiPort:   apiPort,
			ServiceIP: serviceIP,
			Signature: signature,
		},
		DB: &DB{
			DSN: dsn,
		},
		Kafka: &Kafka{
			Broker:              broker,
			KafkaUser:           kafkaUser,
			KafkaPassword:       kafkaPassword,
			DeletedUsersTopic:   deletedUsersTopic,
			BudgetDeleteGroupID: budgetDeleteGroupID,
		},
		SharedRedis: &SharedRedis{
			SharedRedisAddress:  sharedRedisAddress,
			SharedRedisPassword: sharedRedisPassword,
		},
	}
}
