package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/open_db"
	"app/budget-planner/internal/response"
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
	handlerResponse := response.NewHandlerResponse(logger)
	//
	mv := middleware.NewManagerMiddleware(logger, handlerResponse)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(openDb.Postgres)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget)
	//
	budget.NewHandlerBudget(router, serviceBudget, &budget.HandlerBudgetDep{HandlerResponse: handlerResponse, Mv: mv, Logger: logger})
	server := http.Server{
		Addr:    ":" + conf.ApiPort,
		Handler: router,
	}
	errApi := server.ListenAndServe()
	if errApi != nil {
		logger.Error("critical error on the server")
		os.Exit(1)
	}
}
