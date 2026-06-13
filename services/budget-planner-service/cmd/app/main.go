package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/open_db"
	"app/budget-planner/internal/response"
	"app/budget-planner/internal/user"
	"context"
	"net/http"
	"os"
	"shared/loggers"
	"time"
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
	repoExpense := expense.NewRepositoryExpense(openDb.Postgres)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget, repoUser)
	serviceExpense := expense.NewServiceExpense(repoExpense, repoUser, repoBudget)
	//
	router.HandleFunc("GET /health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to write health check: " + errWrite.Error())
		}
	})
	router.HandleFunc("GET /ready", func(writer http.ResponseWriter, request *http.Request) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		sqlDb, errDb := openDb.Postgres.DB.DB()
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
		if status := openDb.Redis.Ping(ctxTimeout); status.Err() != nil {
			logger.Error("ready check failed (Redis ping): " + status.Err().Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("READY")); errWrite != nil {
			logger.Error("failed to write ready check: " + errWrite.Error())
		}
	})
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
