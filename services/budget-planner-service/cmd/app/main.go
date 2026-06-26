package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/finance"
	"context"
	"net/http"
	"os"
	"shared/loggers"
	"shared/open_db"
	"shared/response"
	"shared/shared_middleware"
	"time"
)

func main() {
	logger := loggers.NewLogger()
	//
	conf := budgetconfig.NewConfig(logger)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logger)
	//
	handlerResponse := response.NewHandlerResponse(logger)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, logger, handlerResponse)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(postgres, logger)
	repoExpense := expense.NewRepositoryExpense(postgres, logger)
	repoFinance := finance.NewRepositoryFinance(postgres, logger)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget)
	serviceExpense := expense.NewServiceExpense(repoExpense, serviceBudget)
	serviceFinance := finance.NewServiceFinance(repoFinance, repoBudget, repoExpense)
	//
	router.HandleFunc("GET /health", health(logger))
	router.HandleFunc("GET /ready", ready(postgres, logger))
	budget.NewHandlerBudget(router, serviceBudget, logger, handlerResponse, sharedMv)
	expense.NewHandlerExpense(router, serviceExpense, logger, handlerResponse, sharedMv)
	finance.NewHandlerFinance(router, serviceFinance, handlerResponse, logger, sharedMv)
	chainMv := shared_middleware.Chain(
		sharedMv.Recovery,
		sharedMv.Logging,
	)
	server := http.Server{
		Addr:    ":" + conf.ApiPort,
		Handler: chainMv(router),
	}
	errApi := server.ListenAndServe()
	if errApi != nil {
		logger.Error("critical error on the server: " + errApi.Error())
		os.Exit(1)
	}
}

func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to write health check: ", errWrite)
		}
	}
}
func ready(postgres *open_db.Postgres, logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		sqlDb, errDb := postgres.DB.DB()
		if errDb != nil {
			logger.Error("ready check failed (Postgres init): " + errDb.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if errPing := sqlDb.PingContext(ctxTimeout); errPing != nil {
			logger.Error("ready check failed (Postgres ping): " + errPing.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("READY")); errWrite != nil {
			logger.Error("failed to write ready check: " + errWrite.Error())
		}
	}
}
