package docsconfig

import (
	"os"
	"shared/loggers"

	"github.com/joho/godotenv"
)

type Config struct {
	*Api
	*Services
}
type Api struct {
	ApiPort string
}
type Services struct {
	AuthUserIP        string
	AuthUserPort      string
	BudgetPlannerIP   string
	BudgetPlannerPort string
}

func NewConfig(logger *loggers.Logger) *Config {
	if godotenv.Load() != nil {
		logger.Warn(".env file not found.  This is normal if running inside a container")
		if godotenv.Load(".env.test") != nil {
			logger.Warn(".env.test file not found. This is normal if tests don't run")
		}
	}
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	authUserIP := os.Getenv("AUTH_USER_IP")
	authUserPort := os.Getenv("AUTH_USER_PORT")
	budgetPlannerIP := os.Getenv("BUDGET_PLANNER_IP")
	budgetPlannerPort := os.Getenv("BUDGET_PLANNER_PORT")
	counterEmptyVariables := 0
	if apiPort == "" {
		apiPort = "8080"
		logger.Warn("environment variable 'EXTERNAL_API_PORT' not found. Default value = 8080")
	}
	if authUserIP == "" {
		logger.Error("environment variable 'AUTH_USER_IP' not found")
		counterEmptyVariables++
	}
	if authUserPort == "" {
		logger.Error("environment variable 'AUTH_USER_PORT' not found")
		counterEmptyVariables++
	}
	if budgetPlannerIP == "" {
		logger.Error("environment variable 'BUDGET_PLANNER_IP' not found")
		counterEmptyVariables++
	}
	if budgetPlannerPort == "" {
		logger.Error("environment variable 'BUDGET_PLANNER_PORT' not found")
		counterEmptyVariables++
	}
	if counterEmptyVariables != 0 {
		os.Exit(1)
	}
	return &Config{
		Api: &Api{
			ApiPort: apiPort,
		},
		Services: &Services{
			AuthUserIP:        authUserIP,
			AuthUserPort:      authUserPort,
			BudgetPlannerIP:   budgetPlannerIP,
			BudgetPlannerPort: budgetPlannerPort,
		},
	}
}
