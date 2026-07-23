package main

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/auth"
	"app/auth-service/internal/middleware"
	"app/auth-service/internal/user"
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

// @title           Auth & User Microservice API
// @version         1.0
// @description     Core authentication and identity service with token rotation, profile management, and sliding window rate limiting.
// @host            localhost:8080
// @BasePath        /api/v1
// @tag.name        auth
// @tag.description User registration, authentication, and session management
// @tag.name        user
// @tag.description Profile management, account data retrieval, and user deletion logs
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
	appVariable.KafkaProducer.CloseProducer()

	if errCloseRedis := appVariable.Redis.Close(); errCloseRedis != nil {
		appVariable.Logger.Error("failed to close redis: " + errCloseRedis.Error())
	}
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
	Conf          *authconfig.Config
	KafkaProducer *shared_kafka.KafkaProducer
	SharedRedis   *open_db.Redis
	Redis         *open_db.Redis
	Postgres      *open_db.Postgres
	Logger        *loggers.Logger
	HandlerApp    http.Handler
}

func App() *AppVariable {
	logging := loggers.NewLogger()
	//
	conf := authconfig.NewConfig(logging)
	//
	validate := validator.New()
	//
	responseHandler := response.NewHandlerResponse(logging)
	//
	sharedRedis := open_db.OpenRedis(conf.SharedRedisAddress, conf.SharedRedisPassword, logging)
	rateLimiter := rate_limiter.NewRateLimiter(sharedRedis, logging)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, rateLimiter, logging, responseHandler)
	mv := middleware.NewManagerMiddleware(conf.Signature, logging, responseHandler)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logging)
	redis := open_db.OpenRedis(conf.RedisAddress, conf.RedisPassword, logging)
	kafkaProducer, errInitialProducer := shared_kafka.NewProducer(&shared_kafka.ConfigProducer{
		Brokers:       []string{conf.Broker},
		KafkaUser:     conf.KafkaUser,
		KafkaPassword: conf.KafkaPassword,
		Topic:         conf.DeletedUsersTopic,
	}, logging)
	if errInitialProducer != nil {
		logging.Error("failed to initial producer kafka: " + errInitialProducer.Error())
		os.Exit(1)
	}
	//
	router := http.NewServeMux()
	//
	repoAuth := auth.NewRepositoryRedis(redis, logging)
	repoUser := user.NewRepositoryUser(postgres, logging)
	//
	serviceAuth := auth.NewServiceAuth(repoAuth, repoUser, conf, logging)
	serviceUser := user.NewServiceUser(repoUser, serviceAuth, repoAuth, kafkaProducer, logging)
	//
	router.HandleFunc("GET /health", health(logging))
	router.HandleFunc("GET /ready", ready(postgres, redis, sharedRedis, logging))
	auth.NewHandlerAuth(router, serviceAuth, responseHandler, validate, rateLimiter, logging, mv)
	user.NewHandlerUser(router, serviceUser, responseHandler, validate, logging, mv, sharedMv)
	//
	chainMv := shared_middleware.Chain(
		sharedMv.Logging,
		sharedMv.Recovery,
	)
	return &AppVariable{
		Conf:          conf,
		KafkaProducer: kafkaProducer,
		Logger:        logging,
		HandlerApp:    chainMv(router),
		Postgres:      postgres,
		Redis:         redis,
		SharedRedis:   sharedRedis,
	}
}
func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to writer health check: " + errWrite.Error())
		}
	}
}
func ready(postgres *open_db.Postgres, redis, sharedRedis *open_db.Redis, logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		sqlDb, errDb := postgres.DB.DB()
		if errDb != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Postgres init): " + errDb.Error())
			return
		}
		if errPingPostgres := sqlDb.PingContext(ctxTimeout); errPingPostgres != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Postgres ping): " + errPingPostgres.Error())
			return
		}
		if errPingRedis := redis.Ping(ctxTimeout).Err(); errPingRedis != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Redis ping): " + errPingRedis.Error())
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
			return
		}
	}
}
