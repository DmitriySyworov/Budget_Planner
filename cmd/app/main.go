package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/open_db"
	"app/budget-planner/internal/response"
	"app/budget-planner/internal/user"
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
	repoUser := user.NewRepositoryUser(openDb.Postgres)
	repoBudget := budget.NewRepositoryBudget(openDb.Postgres)
	repoExpense := expense.NewRepositoryExpense(openDb.Postgres)
	//
	serviceUser := user.NewServiceUser(repoUser)
	serviceBudget := budget.NewServiceBudget(repoBudget, repoUser)
	serviceExpense := expense.NewServiceExpense(repoExpense, repoUser, repoBudget)
	//
	user.NewHandlerUser(router, serviceUser)
	budget.NewHandlerBudget(router, serviceBudget, logger, handlerResponse, mv)
	expense.NewHandlerExpense(router, serviceExpense, logger, handlerResponse, mv)
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
