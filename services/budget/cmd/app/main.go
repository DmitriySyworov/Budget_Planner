package main

import (
	budgetconfig "app/budget-planner/config"
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/di"
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/finance"
	"app/budget-planner/internal/listener"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"shared/loggers"
	"shared/ratelimit"
	"shared/response"
	"shared/shkafka"
	"shared/shmiddleware"
	"shared/storage"
	"syscall"
	"time"

	"app/budget-planner/docs"

	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Budget Planner Microservice API
// @version         1.0
// @description     Financial planning and management service. Provides personal budgeting, automated expense tracking, and advanced financial analytics.
// @host            localhost:8081
// @BasePath        /api/v1

// @tag.name        budget
// @tag.description Envelope budgeting, limit configurations, and category allocations
// @tag.name        expense
// @tag.description Daily transaction tracking, merchant categorization, and historic expense logs
// @tag.name        finance
// @tag.description Advanced predictive analytics, financial health scoring, and heavy data aggregation calculations
func main() {
	appVariable := App()
	server := http.Server{
		Addr:    ":" + appVariable.Conf.ApiPort,
		Handler: appVariable.HandlerApp,
	}
	ctxCancel, eventCancel := context.WithCancel(context.Background())
	go appVariable.KafkaConsumer.WaitEvent(ctxCancel, listener.NewKafkaListener(
		appVariable.SharedRedis,
		appVariable.IRepoBudget,
		appVariable.Logger,
	).DeleteDataDeletingUser)
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
	eventCancel()
	appVariable.KafkaConsumer.CloseConsumer()
	appVariable.KafkaProducerExpense.CloseProducer()
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
	Conf                 *budgetconfig.Config
	KafkaConsumer        *shkafka.KafkaConsumer
	KafkaProducerExpense *shkafka.KafkaProducer
	Logger               *loggers.Logger
	HandlerApp           http.Handler
	Postgres             *storage.Postgres
	SharedRedis          *storage.Redis
	ServiceBudget        *budget.ServiceBudget
	IRepoBudget          di.IRepoBudget
}

func App() *AppVariable {
	logging := loggers.NewLogger()
	//
	conf := budgetconfig.NewConfig(logging)
	//
	postgres := storage.OpenPostgres(conf.DSN, logging)
	//
	kafkaConsumer, errInitialConsumer := shkafka.NewConsumer(&shkafka.ConfigConsumer{
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
	kafkaProducerExpense, errInitialProducer := shkafka.NewProducer(&shkafka.ConfigProducer{
		Brokers:       []string{conf.Broker},
		KafkaUser:     conf.KafkaUser,
		KafkaPassword: conf.KafkaPassword,
		Topic:         conf.ExpenseNotificationTopic,
	}, logging)
	if errInitialProducer != nil {
		logging.Error("failed to initial producer kafka: " + errInitialProducer.Error())
		os.Exit(1)
	}
	//
	validate := validator.New()
	//
	responseHandler := response.NewHandlerResponse(logging)
	//
	sharedRedis := storage.OpenRedis(conf.SharedRedisAddress, conf.SharedRedisPassword, logging)
	rateLimiter := ratelimit.NewRateLimiter(sharedRedis, logging)
	//
	sharedMv := shmiddleware.NewManagerSharedMiddleware(conf.Signature, rateLimiter, logging, responseHandler)
	//
	router := http.NewServeMux()
	//
	repoBudget := budget.NewRepositoryBudget(postgres, logging)
	repoExpense := expense.NewRepositoryExpense(postgres, logging)
	repoFinance := finance.NewRepositoryFinance(postgres, logging)
	//
	serviceBudget := budget.NewServiceBudget(repoBudget, logging)
	serviceExpense := expense.NewServiceExpense(repoExpense, serviceBudget, kafkaProducerExpense, logging)
	serviceFinance := finance.NewServiceFinance(repoFinance, repoBudget, repoExpense)
	//
	docs.SwaggerInfo.Host = conf.ServiceIP + ":" + conf.ApiPort
	router.Handle("GET /swagger/{any...}", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	router.HandleFunc("GET /health", health(logging))
	router.HandleFunc("GET /ready", ready(postgres, sharedRedis, logging))
	budget.NewHandlerBudget(router, serviceBudget, logging, responseHandler, validate, sharedMv)
	expense.NewHandlerExpense(router, serviceExpense, logging, responseHandler, validate, sharedMv)
	finance.NewHandlerFinance(router, serviceFinance, responseHandler, logging, sharedMv)
	chainMv := shmiddleware.Chain(
		sharedMv.Logging,
		sharedMv.Recovery,
		sharedMv.RateLimiting,
	)
	return &AppVariable{
		Conf:                 conf,
		KafkaConsumer:        kafkaConsumer,
		KafkaProducerExpense: kafkaProducerExpense,
		Logger:               logging,
		HandlerApp:           chainMv(router),
		Postgres:             postgres,
		SharedRedis:          sharedRedis,
		ServiceBudget:        serviceBudget,
		IRepoBudget:          repoBudget,
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
func ready(postgres *storage.Postgres, sharedRedis *storage.Redis, logger *loggers.Logger) http.HandlerFunc {
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
