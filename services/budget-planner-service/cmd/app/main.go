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
	confApi, logging, handlers := App()
	server := http.Server{
		Addr:    ":" + confApi.ApiPort,
		Handler: handlers,
	}
	errApi := server.ListenAndServe()
	if errApi != nil {
		logging.Error("critical error on the server: " + errApi.Error())
		os.Exit(1)
	}
}
func App() (*budgetconfig.Api, *loggers.Logger, http.Handler) {
	logging := loggers.NewLogger()
	//
	conf := budgetconfig.NewConfig(logging)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logging)
	//
	handlerResponse := response.NewHandlerResponse(logging)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, logging, handlerResponse)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(postgres, logging)
	repoExpense := expense.NewRepositoryExpense(postgres, logging)
	repoFinance := finance.NewRepositoryFinance(postgres, logging)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget)
	serviceExpense := expense.NewServiceExpense(repoExpense, serviceBudget)
	serviceFinance := finance.NewServiceFinance(repoFinance, repoBudget, repoExpense)
	//
	router.HandleFunc("GET /health", health(logging))
	router.HandleFunc("GET /ready", ready(postgres, logging))
	budget.NewHandlerBudget(router, serviceBudget, logging, handlerResponse, sharedMv)
	expense.NewHandlerExpense(router, serviceExpense, logging, handlerResponse, sharedMv)
	finance.NewHandlerFinance(router, serviceFinance, handlerResponse, logging, sharedMv)
	chainMv := shared_middleware.Chain(
		sharedMv.Recovery,
		sharedMv.Logging,
	)
	return conf.Api, logging, chainMv(router)
}
func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to write health check: " + errWrite.Error())
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
