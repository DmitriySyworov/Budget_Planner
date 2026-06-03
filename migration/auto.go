package migration

import (
	"app/budget-planner/internal/loggers"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger := loggers.NewLogger()
	errEnvFile := godotenv.Load()
	if errEnvFile != nil {
		logger.Warn(".env file not found. This is normal if running inside a container")
	}
	dsn := os.Getenv("DSN")
	if dsn == "" {
		logger.Error("environment variable 'DSN' not found")
		os.Exit(1)
	}
	db, errOpen := gorm.Open(postgres.Open(dsn))
	if errOpen != nil {
		logger.Error("failed to connect PostgreSQL")
		os.Exit(1)
	}
	errMigrate := db.AutoMigrate()
	if errMigrate != nil {
		logger.Error("failed to migrate tables")
		os.Exit(1)
	}
}
