package main

import (
	"database/sql"
	"embed"
	"os"
	"shared/loggers"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

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
	db, errOpen := sql.Open("postgres", dsn)
	defer func() {
		if errClose := db.Close(); errClose != nil {
			logger.Error("failed to close sql driver")
		}
	}()
	if errOpen != nil {
		logger.Error("failed to connect PostgreSQL")
		os.Exit(1)
	}
	goose.SetBaseFS(embedMigrations)
	if errDialect := goose.SetDialect("postgres"); errDialect != nil {
		logger.Error("failed to set postgres dialect")
		os.Exit(1)
	}
	if errMigrate := goose.Up(db, "."); errMigrate != nil {
		logger.Error("failed to migrate tables: ", errMigrate)
		os.Exit(1)
	}
}
