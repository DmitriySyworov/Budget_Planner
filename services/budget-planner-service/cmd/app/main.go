package main

import (
	"app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/finance"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"shared/loggers"
	"shared/open_db"
	"shared/rate_limiter"
	"shared/response"
	"shared/shared_kafka"
	"shared/shared_middleware"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
)

func main() {
	appVariable := App()
	server := http.Server{
		Addr:    ":" + appVariable.Conf.ApiPort,
		Handler: appVariable.HandlerApp,
	}
	serverError := make(chan error, 1)
	stopSignal := make(chan os.Signal, 1)
	go func() {
		if errServer := server.ListenAndServe(); errServer != nil && !errors.Is(errServer, http.ErrServerClosed) {
			serverError <- errServer
		}
	}()
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-stopSignal:
		appVariable.Logger.Info("received stop signal, starting graceful shutdown")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if errShutdown := server.Shutdown(ctxShutdown); errShutdown != nil {
			appVariable.Logger.Error("failed graceful shutdown: " + errShutdown.Error())
		}
		cancel()
	case err := <-serverError:
		appVariable.Logger.Error("error on the server: " + err.Error())
	}
	appVariable.EventCancel()
	appVariable.KafkaConsumer.CloseConsumer()
	if errCloseSharedRedis := appVariable.SharedRedis.Close(); errCloseSharedRedis != nil {
		appVariable.Logger.Error("failed to close shared redis: " + errCloseSharedRedis.Error())
	}
	sqlDb, errGetDriverDb := appVariable.Postgres.DB.DB()
	if errGetDriverDb != nil {
		appVariable.Logger.Error("failed to extract sql driver: " + errGetDriverDb.Error())
		return
	}
	if errClosePostgres := sqlDb.Close(); errClosePostgres != nil {
		appVariable.Logger.Error("failed to close Postgres: " + errClosePostgres.Error())
	}
}

type AppVariable struct {
	Conf          *budgetconfig.Config
	KafkaConsumer *shared_kafka.KafkaConsumer
	Logger        *loggers.Logger
	HandlerApp    http.Handler
	EventCancel   context.CancelFunc
	Postgres      *open_db.Postgres
	SharedRedis   *open_db.Redis
}

func App() *AppVariable {
	logging := loggers.NewLogger()
	//
	conf := budgetconfig.NewConfig(logging)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logging)
	kafkaConsumer, errInitialConsumer := shared_kafka.NewConsumer(&shared_kafka.ConfigConsumer{
		Brokers:       []string{conf.Broker},
		KafkaUser:     conf.KafkaUser,
		KafkaPassword: conf.KafkaPassword,
		Topic:         conf.DeletedUsersTopic,
		GroupID:       conf.BudgetDeleteGroupID,
	}, logging)
	if errInitialConsumer != nil {
		logging.Error("failed to initial consumer kafka: " + errInitialConsumer.Error())
		os.Exit(1)
	}
	//
	validate := validator.New()
	//
	responseHandler := response.NewHandlerResponse(logging)
	//
	sharedRedis := open_db.OpenRedis(conf.SharedRedisAddress, conf.SharedRedisPassword, logging)
	rateLimiter := rate_limiter.NewRateLimiter(sharedRedis, logging)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, rateLimiter, logging, responseHandler)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(postgres, logging)
	repoExpense := expense.NewRepositoryExpense(postgres, logging)
	repoFinance := finance.NewRepositoryFinance(postgres, logging)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget, logging)
	serviceExpense := expense.NewServiceExpense(repoExpense, serviceBudget)
	serviceFinance := finance.NewServiceFinance(repoFinance, repoBudget, repoExpense)
	//
	ctxCancel, eventCancel := context.WithCancel(context.Background())
	go kafkaConsumer.WaitEvent(ctxCancel, serviceBudget.DeleteDataDeletingUser)
	//
	router.HandleFunc("GET /health", health(logging))
	router.HandleFunc("GET /ready", ready(postgres, sharedRedis, logging))
	budget.NewHandlerBudget(router, serviceBudget, logging, responseHandler, validate, sharedMv)
	expense.NewHandlerExpense(router, serviceExpense, logging, responseHandler, validate, sharedMv)
	finance.NewHandlerFinance(router, serviceFinance, responseHandler, logging, sharedMv)
	chainMv := shared_middleware.Chain(
		sharedMv.Logging,
		sharedMv.Recovery,
		sharedMv.RateLimiting,
	)
	return &AppVariable{
		Conf:          conf,
		KafkaConsumer: kafkaConsumer,
		Logger:        logging,
		HandlerApp:    chainMv(router),
		EventCancel:   eventCancel,
		Postgres:      postgres,
		SharedRedis:   sharedRedis,
	}
}
func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to write health check: " + errWrite.Error())
		}
	}
}
func ready(postgres *open_db.Postgres, sharedRedis *open_db.Redis, logger *loggers.Logger) http.HandlerFunc {
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
		if errPingSharedRedis := sharedRedis.Ping(ctxTimeout).Err(); errPingSharedRedis != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (shared Redis ping): " + errPingSharedRedis.Error())
			return
		}
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("READY")); errWrite != nil {
			logger.Error("failed to write ready check: " + errWrite.Error())
		}
	}
}
