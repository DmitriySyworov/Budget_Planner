package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/open_db"
	"net/http"
	"os"
)

func main() {
	logger := loggers.NewLogger()
	//
	conf := config.NewConfig(logger)
	//
	openDb := open_db.NewOpenDB(logger, conf.DB)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(openDb)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget)
	//
	budget.NewHandlerBudget(router, serviceBudget)
	server := http.Server{
		Addr:    ":" + conf.ApiPort,
		Handler: router,
	}
	errApi := server.ListenAndServe()
	if errApi != nil {
		logger.Error("critical error on the sever")
		os.Exit(1)
	}
}
